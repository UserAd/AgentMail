package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	// Create a buffer to capture output
	buf := new(bytes.Buffer)

	// Create a new root command for testing
	testRootCmd := &cobra.Command{Use: "agentmail"}
	testVersionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			buf.WriteString("AgentMail version " + GetVersion() + "\n")
		},
	}
	testRootCmd.AddCommand(testVersionCmd)
	testRootCmd.SetOut(buf)
	testRootCmd.SetArgs([]string{"version"})

	// Execute the command
	err := testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check the output
	got := buf.String()
	want := "AgentMail version " + version
	if !strings.Contains(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("GetVersion() returned empty string")
	}
	if v != version {
		t.Errorf("GetVersion() = %q, want %q", v, version)
	}
}
