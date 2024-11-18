package templates

import (
	"fmt"
	"strings"

	"github.com/jabafett/quill/internal/git"
)

// CommitMessagePrompt template for generating commit messages
const CommitMessagePrompt = `Generate a concise commit message for the following git changes:

{{.Diff}}

The commit message should:
1. Follow the Conventional Commits format (type(scope): description)
2. Be specific about what files were changed and how
3. Focus only on the actual changes shown in the diff
4. For deletions, use 'chore' type unless the deletion has a specific purpose
5. Keep the description under 72 characters
6. Include only factual information from the diff
7. No periods or other punctuation at the end of any lines
8. Do not capitalize the first letter of the commit message

Types:
- feat: New features that add functionality (e.g., "feat(auth): add password reset flow")
- fix: Bug fixes or error corrections (e.g., "fix(api): handle null response from server")
- docs: Documentation changes (README, API docs, comments, etc.)
- style: Code style/formatting changes (whitespace, formatting, missing semi-colons)
- refactor: Code changes that neither fix bugs nor add features (restructuring, renaming)
- perf: Performance improvements (e.g., "perf(queries): optimize database indexing")
- test: Adding/modifying tests (unit tests, integration tests, e2e tests)
- chore: Maintenance tasks, dependencies, build changes (no production code change

Files changed: {{.Files}}
Lines added: {{.Added}}
Lines deleted: {{.Deleted}}

Generate only the commit message without any explanation or additional text.`

// GetCommitPrompt formats the commit message template with git diff information
func GetCommitPrompt(repo *git.Repository) (string, error) {
	// Get diff and stats
	diff, err := repo.GetStagedDiff()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	if diff == "" {
		return "", fmt.Errorf("no changes to commit")
	}

	added, deleted, files, err := repo.GetStagedDiffStats()
	if err != nil {
		return "", fmt.Errorf("failed to get diff stats: %w", err)
	}

	// For simple file deletions, return a standard message template
	if len(files) == 1 && added == 0 && deleted > 0 {
		return fmt.Sprintf("chore: remove %s", files[0]), nil
	}

	// Format the prompt with context
	prompt := strings.ReplaceAll(CommitMessagePrompt, "{{.Diff}}", diff)
	prompt = strings.ReplaceAll(prompt, "{{.Files}}", strings.Join(files, ", "))
	prompt = strings.ReplaceAll(prompt, "{{.Added}}", fmt.Sprintf("%d", added))
	prompt = strings.ReplaceAll(prompt, "{{.Deleted}}", fmt.Sprintf("%d", deleted))

	return prompt, nil
}
