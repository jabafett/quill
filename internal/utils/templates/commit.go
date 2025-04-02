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
- Focus only on the actual changes shown in the diff
- Include only factual information from the diff
- If breaking change, add BREAKING CHANGE: in footer
- Do not hallucinate, do not guess

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

<body>

[optional footer(s)]

Example:
feat(auth): add password reset flow

- Add a new password reset flow to the authentication system.
- Include a new route for handling password reset requests.
- Update the login page to display a message indicating that a password reset is required.

BREAKING CHANGE: The password reset flow now requires a confirmation step.

______________________________________________________________________________________________________________________

<files>
Files changed: {{.Files}}
</files>
<diff>
{{.Diff}}
</diff>
Generate only the commit message without any explanation or additional text.
`
)
