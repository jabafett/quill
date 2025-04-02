package context

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// FileTypeDetector handles detection of file types
type FileTypeDetector struct{}

// DetectFileType determines the type of file based on content and path
func (d *FileTypeDetector) DetectFileType(path string, content []byte) string {
	// Try extension-based detection first
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".rs":
		return "rust"
	case ".c", ".cpp", ".h", ".hpp":
		return "cpp"
	case ".md", ".markdown":
		return "markdown"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".sql":
		return "sql"
	}

	// Try to detect by content patterns if no extension match
	// python shebang
	if bytes.Contains(content, []byte("#!/usr/bin/env python")) {
		return "python"
	}
	// go shebang
	if bytes.Contains(content, []byte("#!/usr/bin/env go")) {
		return "go"
	}
	// rust shebang
	if bytes.Contains(content, []byte("#!/usr/bin/env rust")) {
		return "rust"
	}
	// cpp shebang
	if bytes.Contains(content, []byte("#!/usr/bin/env cpp")) {
		return "cpp"
	}
	if bytes.Contains(content, []byte("fn ")) || bytes.Contains(content, []byte("use std::")) {
		return "rust"
	}
	if bytes.Contains(content, []byte("#include")) || bytes.Contains(content, []byte("namespace ")) {
		return "cpp"
	}
	if bytes.Contains(content, []byte("public class")) || bytes.Contains(content, []byte("package ")) {
		return "java"
	}
	if bytes.Contains(content, []byte("impl ")) {
		return "rust"
	}
	if bytes.Contains(content, []byte("class ")) || bytes.Contains(content, []byte("template<")) {
		return "cpp"
	}

	// Fallback to mime type detection
	mtype := mimetype.Detect(content)
	switch mtype.String() {
	case "application/x-golang":
		return "go"
	case "application/javascript", "text/javascript":
		return "javascript"
	case "text/typescript":
		return "typescript"
	case "text/x-python":
		return "python"
	case "text/x-java":
		return "java"
	case "text/x-ruby":
		return "ruby"
	case "text/x-php":
		return "php"
	case "text/x-rust":
		return "rust"
	case "text/x-c", "text/x-c++":
		return "cpp"
	}

	return strings.Split(mtype.String(), ";")[0]
}
