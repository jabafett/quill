package context

import (
	"bytes"
	"path/filepath"

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
	case ".ts", ".tsx":
		return "typescript"
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
		return "c++"
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

	// Fallback to mime type detection
	mtype := mimetype.Detect(content)
	// Map mime types to our internal type system
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
		return "c++"
	}
	// Try to detect by content patterns
	if bytes.Contains(content, []byte("<?php")) {
		return "php"
	}
	if bytes.Contains(content, []byte("#!/usr/bin/env python")) {
		return "python"
	}
	if bytes.Contains(content, []byte("package ")) && bytes.Contains(content, []byte("import ")) {
		return "go"
	}

	// If all else fails, return the mime type
	return mtype.String()
}
