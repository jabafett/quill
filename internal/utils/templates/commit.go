package templates

// Template definitions
const (
	CommitMessageTemplate = `Your task is to generate a commit message.
The commit message should:
- Separate subject from body with a blank line
- Limit the subject line to 72 characters
- Use imperative mood ("add" not "added", "change" not "changed")
- Use body to explain what and why things were changed/added not how
- Be specific about what was changed and how
- Do not capitalize the first letter of the commit message
- No periods or other punctuation at the end of any lines
- The scope should not be the name of a file
- Focus only on the actual changes shown in the diff
- Include only factual information from the diff
- If breaking change, add BREAKING CHANGE: in footer
- Reference issues and PRs in footer when applicable
- The description, body, nor the footer should contain the filename or path

Types:
feat: New features that add functionality (e.g., "feat(auth): add password reset flow")
fix: Bug fixes or error corrections (e.g., "fix(api): handle null response from server")
docs: Documentation changes (README, API docs, comments, etc.)
style: Code style/formatting changes (whitespace, formatting, missing semi-colons)
refactor: Code changes that neither fix bugs nor add features (restructuring, renaming)
perf: Performance improvements (e.g., "perf(queries): optimize database indexing")
test: Adding/modifying tests (unit tests, integration tests, e2e tests)
chore: Maintenance tasks, dependencies, build changes (no production code change)

Template:
<type>(<scope>): <description>

- subject body

[optional footer(s)]

Context:
<repo_context>
<files>
Files changed: {{.Files}}
</files>
</repo_context>
<diff>
{{.Diff}}
</diff>
Generate only the commit message without any explanation or additional text.
`
) 