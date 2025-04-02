package context_test

import (
	"testing"

	"github.com/jabafett/quill/internal/utils/context"
)

func TestFileTypeDetector(t *testing.T) {
	detector := &context.FileTypeDetector{}

	tests := []struct {
		name     string
		path     string
		content  []byte
		wantType string
	}{
		{
			name:     "Go file by extension",
			path:     "test.go",
			content:  []byte(`package main`),
			wantType: "go",
		},
		{
			name:     "JavaScript file by extension",
			path:     "test.js",
			content:  []byte(`console.log("test")`),
			wantType: "javascript",
		},
		{
			name:     "TypeScript file by extension",
			path:     "test.ts",
			content:  []byte(`interface Test {}`),
			wantType: "typescript",
		},
		{
			name:     "Python file by extension",
			path:     "test.py",
			content:  []byte(`print("test")`),
			wantType: "python",
		},
		{
			name:     "Java file by extension",
			path:     "test.java",
			content:  []byte(`public class Test {}`),
			wantType: "java",
		},
		{
			name:     "Ruby file by extension",
			path:     "test.rb",
			content:  []byte(`puts "test"`),
			wantType: "ruby",
		},
		{
			name:     "PHP file by extension",
			path:     "test.php",
			content:  []byte(`<?php echo "test"; ?>`),
			wantType: "php",
		},
		{
			name:     "Rust file by extension",
			path:     "test.rs",
			content:  []byte(`fn main() {}`),
			wantType: "rust",
		},
		{
			name:     "C++ header file by extension",
			path:     "test.hpp",
			content:  []byte(`class Test {};`),
			wantType: "cpp",
		},
		{
			name:     "Markdown file by extension",
			path:     "test.md",
			content:  []byte(`# Test`),
			wantType: "markdown",
		},
		{
			name:     "YAML file by extension",
			path:     "test.yml",
			content:  []byte(`key: value`),
			wantType: "yaml",
		},
		{
			name:     "JSON file by extension",
			path:     "test.json",
			content:  []byte(`{"key": "value"}`),
			wantType: "json",
		},
		{
			name:     "XML file by extension",
			path:     "test.xml",
			content:  []byte(`<?xml version="1.0"?>`),
			wantType: "xml",
		},
		{
			name:     "SQL file by extension",
			path:     "test.sql",
			content:  []byte(`SELECT * FROM test;`),
			wantType: "sql",
		},
		// Content-based detection tests
		{
			name:     "PHP file by content",
			path:     "test.txt",
			content:  []byte(`<?php echo "test"; ?>`),
			wantType: "php",
		},
		{
			name:     "Python file by shebang",
			path:     "test",
			content:  []byte(`#!/usr/bin/env python\nprint("test")`),
			wantType: "python",
		},
		// Unknown file type
		{
			name:     "Unknown file type",
			path:     "test.unknown",
			content:  []byte(`some random content`),
			wantType: "text/plain", // Default mime type for text content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType := detector.DetectFileType(tt.path, tt.content)
			if gotType != tt.wantType {
				t.Errorf("DetectFileType() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}
