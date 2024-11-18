//Files: file list, FilesContents: file contents, AllTheFilesClarification: clarification if all the files content is included, GitHistory: git history, GitHistoryClarification: clarification if the whole git history is included
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

Provide a concise summary of:
1. The repository language and dependencies
2. The repository's architecture and design patterns
3. Any other relevant context about what the repository is for` 