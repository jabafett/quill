package providers

import (
        "context"
        "fmt"
        "sync"

        "github.com/jabafett/quill/internal/factories"
        "github.com/jabafett/quill/internal/utils/ai"
        "github.com/jabafett/quill/internal/utils/config"
        "github.com/jabafett/quill/internal/utils/debug"
        "github.com/jabafett/quill/internal/utils/git"
        "github.com/jabafett/quill/internal/utils/helpers"
)

// SuggestFactory handles the generation of commit grouping suggestions
type SuggestFactory struct {
        config          *config.Config
        repo            *git.Repository
        templates       *factories.TemplateFactory
        provider        factories.Provider
        contextProvider *factories.ContextProvider
        stagedOnly      bool
        unstagedOnly    bool
}

// NewSuggestFactory creates a new factory specifically for the suggest command
func NewSuggestFactory(opts factories.ProviderOptions) (*SuggestFactory, error) {
        var (
                cfg       *config.Config
                repo      *git.Repository
                templates *factories.TemplateFactory
                provider  factories.Provider
                errChan   = make(chan error, 4)
                wg        sync.WaitGroup
        )

        debug.Log("Starting suggest factory")

        // Load components concurrently
        wg.Add(3)

        // Load configuration
        go func() {
                defer wg.Done()
                var err error
                cfg, err = config.LoadConfig()
                if err != nil {
                        errChan <- fmt.Errorf("failed to load config: %w", err)
                }
                debug.Log("Finished loading configuration")
        }()

        // Initialize git object
        go func() {
                defer wg.Done()
                var err error
                repo, err = git.NewRepository(".")
                if err != nil {
                        errChan <- err
                }
                debug.Log("Finished initializing git repository")
        }()

        // Initialize template factory
        go func() {
                defer wg.Done()
                var err error
                templates, err = factories.NewTemplateFactory()
                if err != nil {
                        errChan <- fmt.Errorf("failed to create template factory: %w", err)
                }
                debug.Log("Finished initializing template factory")
        }()

        // Wait for config to be loaded before creating provider
        wg.Wait()
        close(errChan)

        // Check for any errors
        for err := range errChan {
                return nil, err
        }

        // Create provider with the loaded config
        provider, err := factories.NewProvider(cfg, opts)
        if err != nil {
                return nil, fmt.Errorf("failed to create provider: %w", err)
        }

        debug.Log("Finished creating provider")

        // Initialize context provider
        repoRootPath, err := repo.GetRepoRootPath()
        if err != nil {
                return nil, fmt.Errorf("failed to get repo root path: %w", err)
        }

        contextProvider, err := factories.NewContextProvider(
                factories.WithRepoRootPath(repoRootPath),
        )
        if err != nil {
                debug.Log("Warning: Failed to create context provider: %v. Proceeding without repository context.", err)
        }

        factory := &SuggestFactory{
                config:          cfg,
                repo:            repo,
                templates:       templates,
                provider:        provider,
                contextProvider: contextProvider,
                stagedOnly:      opts.StagedOnly,
                unstagedOnly:    opts.UnstagedOnly,
        }

        return factory, nil
}

// Suggest generates commit grouping suggestions based on changes
func (f *SuggestFactory) Suggest(ctx context.Context) ([]helpers.SuggestionGroup, error) {
        // Check for changes
        hasStagedChanges, _ := f.repo.HasStagedChangesOptimized()
        
        // Get staged diff if needed
        var stagedDiff string
        var stagedFiles []string
        var err error
        
        if !f.unstagedOnly {
                if hasStagedChanges {
                        stagedDiff, err = f.repo.GetStagedDiffOptimized()
                        if err != nil {
                                return nil, fmt.Errorf("failed to get staged diff: %w", err)
                        }
                        
                        stagedFiles, err = f.repo.GetStagedFilesOptimized()
                        if err != nil {
                                return nil, fmt.Errorf("failed to get staged files: %w", err)
                        }
                }
        }
        
        // Get unstaged diff if needed
        var unstagedDiff string
        var unstagedFiles []string
        
        if !f.stagedOnly {
                // Use git diff to get unstaged changes
                cmd := "git diff --no-color"
                output, err := helpers.ExecuteCommand(cmd)
                if err != nil {
                        return nil, fmt.Errorf("failed to get unstaged diff: %w", err)
                }
                unstagedDiff = output
                
                // Get unstaged files
                cmd = "git diff --name-only"
                output, err = helpers.ExecuteCommand(cmd)
                if err != nil {
                        return nil, fmt.Errorf("failed to get unstaged files: %w", err)
                }
                if output != "" {
                        unstagedFiles = helpers.SplitLines(output)
                }
        }
        
        // Get untracked files that are not gitignored
        untrackedFiles, err := f.repo.GetUntrackedFiles()
        if err != nil {
                debug.Log("Warning: Failed to get untracked files: %v", err)
                untrackedFiles = []string{}
        }
        
        // Get content of untracked files
        untrackedContent := ""
        if len(untrackedFiles) > 0 {
                debug.Log("Found %d untracked files to include in context", len(untrackedFiles))
                var untrackedContentBuilder strings.Builder
                
                for _, file := range untrackedFiles {
                        content, err := f.repo.GetFileContent(file)
                        if err != nil {
                                debug.Log("Warning: Failed to read content of untracked file %s: %v", file, err)
                                continue
                        }
                        
                        untrackedContentBuilder.WriteString(fmt.Sprintf("File: %s\n", file))
                        untrackedContentBuilder.WriteString(content)
                        untrackedContentBuilder.WriteString("\n\n")
                }
                
                untrackedContent = untrackedContentBuilder.String()
                
                // Add untracked files to the list of files to consider
                unstagedFiles = append(unstagedFiles, untrackedFiles...)
        }
        
        // Check if we have any changes to work with
        if (f.stagedOnly && !hasStagedChanges) || 
           (f.unstagedOnly && len(unstagedFiles) == 0) ||
           (!f.stagedOnly && !f.unstagedOnly && !hasStagedChanges && len(unstagedFiles) == 0 && len(untrackedFiles) == 0) {
                return nil, helpers.ErrNoChanges{}
        }
        
        // Get repository context if available
        repoContext := ""
        if f.contextProvider != nil && f.contextProvider.HasSummary() {
                repoContext = f.contextProvider.GetRepoSummary()
                debug.Log("Added repository summary to prompt data")
        } else {
                debug.Log("No repository summary available. Run 'quill index' first for context-aware suggestions.")
        }

        // Prepare template data
        data := map[string]interface{}{
                "Context":          repoContext,
                "Staged":           stagedDiff,
                "Unstaged":         unstagedDiff,
                "Untracked":        untrackedContent,
                "UntrackedFiles":   untrackedFiles,
        }

        // Generate prompt from template
        prompt, err := f.templates.Generate(factories.SuggestionType, data)
        if err != nil {
                return nil, fmt.Errorf("failed to generate suggestion prompt: %w", err)
        }

        // Generate suggestions using the AI provider
        opts := ai.GenerateOptions{
                MaxCandidates: f.config.Core.DefaultCandidates,
        }
        if temp := f.config.Providers[f.config.Core.DefaultProvider].Temperature; temp > 0 {
                opts.Temperature = &temp
        }
        
        debug.Log("Sending prompt to AI provider for suggestions")
        responses, err := f.provider.Generate(ctx, prompt, opts)
        if err != nil {
                return nil, fmt.Errorf("failed to generate suggestions: %w", err)
        }
        
        // Parse the responses into suggestion groups
        suggestions := make([]helpers.SuggestionGroup, 0, len(responses))
        
        for i, response := range responses {
                // Parse the AI response into structured suggestions
                allFiles := append(append([]string{}, stagedFiles...), unstagedFiles...)
                groups := helpers.ParseSuggestionResponse(response, stagedFiles, allFiles)
                
                // Add each group to our suggestions
                for j, group := range groups {
                        group.ID = fmt.Sprintf("suggestion-%d-%d", i+1, j+1)
                        suggestions = append(suggestions, group)
                }
        }
        
        return suggestions, nil
}

