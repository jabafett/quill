// Staging template for generating a suite of commit groupings along with commit messages
package templates

// SuggestTemplate defines the template for generating a suite of commit groupings along with descriptions
const SuggestTemplate = `# TASK
Analyze the repository changes and suggest logical commit groupings. Group related changes together and create appropriate conventional commit messages for each group.

# GROUPING GUIDELINES
- Group changes by functionality, not by file type
- Keep related changes together in a single commit
- Include tests with the implementation they test
- Include documentation with the code it documents
- If all changes are related to a single feature or fix, use just one grouping

## Commit Rules
- Separate subject from body with a blank line
- Limit the subject line to 72 characters
- Use imperative mood ("add" not "added", "change" not "changed")
- Do not capitalize the first letter of the commit message
- No periods or other punctuation at the end of any lines
- Use body to explain what and why things were changed/added not how
- If breaking change, add BREAKING CHANGE: in footer
- Be specific about what was changed and how
- Please refrain from discussing formatting changes nor inferences about the scope of the change through code that has only been reformatted (e.g., indentation, line length, etc.) look for functional changes, sometimes autoformatters will change many lines and this is not relevant but code might have still changed within the reformatted code
- Sift through the noise in the diff and information provided to zero in on what was modified, added, or removed

# COMMIT MESSAGE GUIDELINES

## Format
<type>(<scope>): <description>

<body>

[optional footer(s)]

## Types
- feat: New features that add functionality
- fix: Bug fixes or error corrections
- docs: Documentation changes
- style: Code style/formatting changes (whitespace, formatting, etc.)
- refactor: Code changes that neither fix bugs nor add features
- perf: Performance improvements
- test: Adding/modifying tests
- chore: Maintenance tasks, dependencies, build changes

# RESPONSE FORMAT

Your response must be formatted in XML as follows:


<suggestions>
  <group>
    <description>Brief description of the first logical grouping</description>
    <files>
      <file>file1.ext</file>
      <file>file2.ext</file>
      <file>file3.ext</file>
    </files>
    <commit>
      <header>type(scope): short description</header>
      <body>Detailed explanation of what was changed and why</body>
      <footer>Optional footer notes like BREAKING CHANGE</footer>
    </commit>
  </group>
  <group>
    <description>Brief description of the second logical grouping</description>
    <files>
      <file>file4.ext</file>
      <file>file5.ext</file>
    </files>
    <commit>
      <header>type(scope): short description</header>
      <body>Detailed explanation of what was changed and why</body>
    </commit>
  </group>
</suggestions>

# EXAMPLE

<suggestions>
  <group>
    <description>Authentication system implementation</description>
    <files>
      <file>auth/login.go</file>
      <file>auth/middleware.go</file>
      <file>auth/user.go</file>
      <file>tests/auth_test.go</file>
    </files>
    <commit>
      <header>feat(auth): implement user authentication system</header>
      <body>- Add a new password reset flow to the authentication system
- Include a new route for handling password reset requests
- Update the login page to display a message indicating that a password reset is required</body>
      <footer>BREAKING CHANGE: The password reset flow now requires a confirmation step</footer>
    </commit>
  </group>
  <group>
    <description>Fix database connection timeout</description>
    <files>
      <file>db/connection.go</file>
      <file>config/database.yaml</file>
    </files>
    <commit>
      <header>fix(db): increase connection timeout to prevent disconnects</header>
      <body>- Increase the database connection timeout from 5s to 30s to prevent disconnects during high load periods
	- Add a new password reset flow to the authentication system.
	- Include a new route for handling password reset requests.
	- Update the login page to display a message indicating that a password reset is required.
    </commit>
  </group>
</suggestions>

# REPOSITORY CONTEXT
{{.Context}}

# CHANGES TO ANALYZE

## Staged Changes
{{.Staged}}

## Unstaged Changes
{{.Unstaged}}

## Untracked Files
{{.Untracked}}



`
