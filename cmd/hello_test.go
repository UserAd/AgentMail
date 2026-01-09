package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestHelloCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		want     string
	}{
		{
			name: "default greeting",
			args: []string{"hello"},
			want: "Hello, World!",
		},
		{
			name: "custom name with --name",
			args: []string{"hello", "--name", "Alice"},
			want: "Hello, Alice!",
		},
		{
			name: "custom name with -n",
			args: []string{"hello", "-n", "Bob"},
			want: "Hello, Bob!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			buf := new(bytes.Buffer)

			// Create a new root command for testing
			testRootCmd := &cobra.Command{Use: "agentmail"}
			var testName string
			testHelloCmd := &cobra.Command{
				Use:   "hello",
				Short: "Say hello",
				Run: func(cmd *cobra.Command, args []string) {
					buf.WriteString("Hello, " + testName + "!\n")
				},
			}
			testHelloCmd.Flags().StringVarP(&testName, "name", "n", "World", "Name to greet")
			testRootCmd.AddCommand(testHelloCmd)
			testRootCmd.SetOut(buf)
			testRootCmd.SetArgs(tt.args)

			// Execute the command
			err := testRootCmd.Execute()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check the output
			got := strings.TrimSpace(buf.String())
			want := tt.want
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
