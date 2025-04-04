// Staging template for generating a suite of commit groupings along with commmit messages
package templates

// SuggestTemplate defines the template for generating a suite of commit groupings along with descriptions
const SuggestTemplate = `Based on the following analysis:

Repository Context:
{{.Context}}

Staged diff:
{{.Staged}}

Unstaged diff:
{{.Unstaged}}

Analyze the changes and suggest logical groupings for commits. For each grouping:
1. Provide a clear description of what the grouping represents
2. List the specific files that should be included
3. Suggest a conventional commit message (e.g., feat(scope): description)
4. Indicate version impact if applicable (major/minor/patch)

Format your response as follows for each grouping:

Group 1: [Brief description of the first logical grouping]
Files:
- [file1.ext]
- [file2.ext]
- [file3.ext]
Commit message: [conventional commit message]
Impact: [major/minor/patch]

Group 2: [Brief description of the second logical grouping]
Files:
- [file4.ext]
- [file5.ext]
Commit message: [conventional commit message]
Impact: [major/minor/patch]

Provide 2-3 logical groupings that make sense based on the changes.
`