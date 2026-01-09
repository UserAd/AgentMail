package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Create a buffer to capture output
	buf := new(bytes.Buffer)

	// Reset rootCmd for testing
	testCmd := rootCmd
	testCmd.SetOut(buf)
	testCmd.SetErr(buf)
	testCmd.SetArgs([]string{"--help"})

	// Execute the command
	err := testCmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check the output contains expected text
	got := buf.String()
	if got == "" {
		t.Error("Expected help output, got empty string")
	}
}
