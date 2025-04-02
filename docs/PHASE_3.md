# Phase 3: Core Intelligence Implementation

## Overview

Phase 3 focuses on building the core intelligence features: indexing repository context and using that context to enhance commit message generation (`generate`) and enable smart change suggestions (`suggest`).

## Core Tasks

### 1. Context Indexing (`index` command)

*   **Implement `index` Command:**
    *   Implement `internal/cmd/index.go`.
    *   Implement `internal/providers/index_provider.go`.
    *   Integrate with `ContextEngine` (`internal/factories/context_factory.go`) and any other needed factory, to analyze repository files based on configuration (e.g., respecting `.gitignore`).
*   **Repository-wide Context Persistence:**
    *   Define strategy for storing the aggregated `RepositoryContext` (defined in `internal/utils/context/types.go`).
    *   Leverage the existing Badger cache (`internal/utils/cache/cache.go`) for persistent storage of the indexed context.
    *   Design schema/key structure for efficient retrieval of the full or partial context.
*   **Indexing Efficiency:**
    *   Implement logic for incremental updates to avoid re-analyzing unchanged files on subsequent runs.
    *   Optimize for performance and memory usage, especially in large repositories.
    *   Develop cache invalidation strategy (e.g., based on file modifications).

### 2. Context-Aware Generation (`generate` command enhancement)

*   **Integrate Indexed Context:**
    *   Modify `internal/factories/generate.go` to optionally load the persisted `RepositoryContext` built by the `index` command.
    *   Update the `CommitMessageTemplate` (`internal/utils/templates/commit.go`) data structure and prompt to incorporate relevant context (e.g., related symbols, dependencies, file types) alongside the diff.
*   **Refine AI Prompting:**
    *   Adjust the `CommitMessageTemplate` prompt to guide the AI in utilizing the richer context for more accurate type, scope, and description generation.
    *   Experiment with including summaries of related changed symbols or affected dependencies.

### 3. Smart Suggestions (`suggest` command)

*   **Implement `suggest` Command:**
    *   Create `internal/cmd/suggest.go`.
    *   Load the persisted `RepositoryContext`.
    *   Analyze staged and potentially unstaged changes (`internal/utils/git/git.go`) against the indexed context.
*   **Suggestion Logic:**
    *   Develop algorithms or heuristics to identify logically related changes based on shared symbols, dependencies, file relationships, or change types derived from the context.
    *   Integrate with AI providers (`internal/factories/provider_factory.go`) using `SuggestTemplate` (`internal/utils/templates/suggest.go`) to generate:
        *   Recommended groupings of files for commits.
        *   Potential commit messages for suggested groups.
        *   Identification of related but unstaged files.
        *   Basic impact analysis (e.g., breaking change indicators).
*   **User Interface:**
    *   Design and implement UI elements (potentially extending `internal/ui/`) to present suggestions clearly to the user for review and action.

## Success Criteria

*   `quill index` successfully analyzes and persists repository context efficiently.
*   `quill generate` produces more contextually relevant commit messages when an index is present.
*   `quill suggest` provides logical and actionable commit grouping recommendations based on the indexed context.