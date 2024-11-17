package tests

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/jabafett/quill/internal/debug"
)

func TestDebugOutput(t *testing.T) {
	// Capture stderr output
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Initialize debug mode
	debug.Initialize(true)

	// Test debug logging
	testMessage := "test debug message"
	debug.Log(testMessage)

	// Test debug dump
	testVar := struct{ Name string }{"test"}
	debug.Dump("testVar", testVar)

	// Restore stderr
	w.Close()
	os.Stderr = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify debug message
	if !strings.Contains(output, "DEBUG: "+testMessage) {
		t.Errorf("Expected debug output to contain message: %s", testMessage)
	}

	// Verify dump output
	if !strings.Contains(output, "DEBUG: testVar = ") {
		t.Error("Expected debug output to contain variable dump")
	}

	// Test debug disabled
	debug.Initialize(false)
	if debug.IsDebug() {
		t.Error("Expected debug mode to be disabled")
	}
} 