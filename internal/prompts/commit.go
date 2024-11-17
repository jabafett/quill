package prompts

import (
	"fmt"
)

const (
	// base template for generating commit messages
	CommitMessageTemplate = `Analyze the following git diff and generate a conventional commit message.
The message should follow this format:
<type>(<scope>): <description>

[optional body]

[optional footer(s)]

Types:
- feat: New features that add functionality (e.g., "feat(auth): add password reset flow")
- fix: Bug fixes or error corrections (e.g., "fix(api): handle null response from server")
- docs: Documentation changes (README, API docs, comments, etc.)
- style: Code style/formatting changes (whitespace, formatting, missing semi-colons)
- refactor: Code changes that neither fix bugs nor add features (restructuring, renaming)
- perf: Performance improvements (e.g., "perf(queries): optimize database indexing")
- test: Adding/modifying tests (unit tests, integration tests, e2e tests)
- chore: Maintenance tasks, dependencies, build changes (no production code change)

Rules:
- Use imperative mood ("add" not "added", "change" not "changed")
- Don't capitalize first letter of description
- No period at the end of the description
- Keep first line under 72 characters
- Separate subject from body with a blank line
- Use body to explain what and why vs. how
- If breaking change, add BREAKING CHANGE: in footer
- Reference issues and PRs in footer when applicable

Example:
feat(auth): add password reset flow

Git diff:
%s`

	// used by the suggest command to group related files
	SuggestGroupsTemplate = `Analyze these changed files and suggest logical groupings for separate commits.

Rules for grouping:
- Group files by related functionality or purpose
- Consider directory structure and file types
- Separate unrelated changes into different groups
- Identify breaking changes or major refactors
- Keep test files with their implementation files
- Group documentation changes separately
- Consider dependency changes as separate groups
- Always group .gitignore changes separately

For each group, provide:
1. A descriptive name for the group
2. List of files in the group
3. Suggested commit type (feat, fix, etc.)
4. The conventional commit message for the group

Output format:
Group 1: [name]
Type: [commit type]
Files:
- [file1]
- [file2]
Commit message: [commit message]

Group 2: [name]
...

Rules for commit messages:
Types:
- feat: New features that add functionality (e.g., "feat(auth): add password reset flow")
- fix: Bug fixes or error corrections (e.g., "fix(api): handle null response from server")
- docs: Documentation changes (README, API docs, comments, etc.)
- style: Code style/formatting changes (whitespace, formatting, missing semi-colons)
- refactor: Code changes that neither fix bugs nor add features (restructuring, renaming)
- perf: Performance improvements (e.g., "perf(queries): optimize database indexing")
- test: Adding/modifying tests (unit tests, integration tests, e2e tests)
- chore: Maintenance tasks, dependencies, build changes (no production code change)

Rules:
- Use imperative mood ("add" not "added", "change" not "changed")
- Don't capitalize first letter of description
- No period at the end of the description
- Keep first line under 72 characters
- Separate subject from body with a blank line
- Use body to explain what and why vs. how
- If breaking change, add BREAKING CHANGE: in footer
- Reference issues and PRs in footer when applicable

Example:
feat(auth): add password reset flow

Changed files:
%s
Git diff:
%s`
)

// GetCommitPrompt formats the commit message template with the provided diff
func GetCommitPrompt(diff string) string {
	return fmt.Sprintf(CommitMessageTemplate, diff)
}
