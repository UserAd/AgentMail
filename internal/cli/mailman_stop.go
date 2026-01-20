package cli

import (
	"fmt"
	"io"
	"os"

	"agentmail/internal/daemon"
	"agentmail/internal/mail"
)

// MailmanStopOptions configures the MailmanStop command behavior.
type MailmanStopOptions struct {
	RepoRoot string // Repository root (defaults to finding git root)
}

// MailmanStop implements the agentmail mailman stop command.
// Creates a .stop file to signal the daemon to shut down.
//
// The function attempts to find the repository root via FindGitRoot,
// falling back to os.Getwd if not in a git repository. If both fail,
// repoRoot will be empty and file creation will fail with a clear error.
//
// Exit codes:
// - 0: Success (stop signal sent)
// - 1: Error (file exists or filesystem error)
func MailmanStop(stdout, stderr io.Writer, opts MailmanStopOptions) int {
	// Find repository root
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = mail.FindGitRoot()
		if err != nil {
			repoRoot, _ = os.Getwd()
		}
	}

	stopPath := daemon.StopFilePath(repoRoot)

	// Atomic create - fails if file exists (O_CREATE|O_EXCL)
	f, err := os.OpenFile(stopPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600) // #nosec G304 - stopPath is constructed from constants
	if err != nil {
		if os.IsExist(err) {
			fmt.Fprintln(stderr, "Stop already pending")
			return 1
		}
		fmt.Fprintf(stderr, "Failed to send stop signal: %v\n", err)
		return 1
	}
	_ = f.Close() // G104: file was just created successfully, close error is non-critical

	fmt.Fprintln(stdout, "Stop signal sent")
	return 0
}
