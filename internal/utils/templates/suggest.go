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

Provide suggestions for:
1. Additional files that should be included
2. Additional commit grouping recommendations
3. Additional commit descriptions
4. Version impact if applicable (major/minor/patch)

Response Template:
`