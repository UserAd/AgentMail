package cli

import (
	"os"
)

// IsStdinPipe returns true if stdin is a pipe or redirect (not a terminal).
// This is used to detect when input is being piped into the command,
// allowing the message to be read from stdin instead of command arguments.
func IsStdinPipe() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
