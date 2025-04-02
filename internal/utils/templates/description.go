// use this file as inspiration to create a context summary prompt to fill in the file data field in the RepositoryContext's Description field!
package templates

// ContextExtractionTemplate defines the template for extracting context from repository
const ContextExtractionTemplate = `Analyze the following repository context:
Files:
{{range .Files}}
- {{.}}
{{end}}

File Contents:
{{range $file, $content := .FilesContent}}
=== {{$file}} ===
{{$content}}
{{end}}
{{.AllTheFilesClarification}}

Git history:
{{.GitHistory}}
{{.GitHistoryClarification}}

Provide a concise summary of (but not limited to):
1. The repository's purpose and/or goal
2. The repository's architecture and design patterns
3. Any other relevant context about what the repository is for or does` 