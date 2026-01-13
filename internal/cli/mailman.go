package cli

import (
	"io"
	"os"

	"agentmail/internal/daemon"
	"agentmail/internal/mail"
)

// MailmanOptions configures the Mailman command behavior.
// Used for testing to mock file system operations.
type MailmanOptions struct {
	Daemonize bool   // Run in background (--daemon flag)
	RepoRoot  string // Repository root (defaults to finding git root)
}

// Mailman implements the agentmail mailman command.
// Starts the mailman daemon in foreground or background mode.
//
// Exit codes:
// - 0: Success
// - 1: Error (failed to start)
// - 2: Daemon already running
func Mailman(stdout, stderr io.Writer, opts MailmanOptions) int {
	// Find repository root if not provided
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = mail.FindGitRoot()
		if err != nil {
			// Not in a git repository
			// For mailman, we need .agentmail/ to store PID file
			// Fall back to current directory
			repoRoot, err = os.Getwd()
			if err != nil {
				return 1
			}
		}
	}

	// Check if this is a daemon child process
	if daemon.IsDaemonChild() {
		// Child process: run in foreground mode regardless of opts.Daemonize
		return daemon.StartDaemon(repoRoot, false, stdout, stderr)
	}

	// Start daemon (foreground or background based on opts.Daemonize)
	return daemon.StartDaemon(repoRoot, opts.Daemonize, stdout, stderr)
}
