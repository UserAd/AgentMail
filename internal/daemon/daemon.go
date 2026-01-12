// Package daemon provides functionality for the mailman daemon process.
// It handles PID file management and daemon startup/shutdown.
package daemon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"agentmail/internal/mail"
)

// PIDFile is the filename for the mailman daemon PID file within .git/mail/
const PIDFile = "mailman.pid"

// PIDFilePath returns the full path to the PID file for a given repository root.
func PIDFilePath(repoRoot string) string {
	return filepath.Join(repoRoot, mail.MailDir, PIDFile)
}

// ReadPID reads the PID from the mailman.pid file.
// Returns 0 if the file doesn't exist (not an error).
// Returns an error if the file exists but contains invalid content.
func ReadPID(repoRoot string) (int, error) {
	pidPath := PIDFilePath(repoRoot)

	content, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	// Parse PID from file content (strip newline)
	pidStr := strings.TrimSpace(string(content))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID file content: %w", err)
	}

	return pid, nil
}

// WritePID writes the given PID to the mailman.pid file.
// Creates the file if it doesn't exist, overwrites if it does.
func WritePID(repoRoot string, pid int) error {
	// Ensure mail directory exists
	if err := mail.EnsureMailDir(repoRoot); err != nil {
		return fmt.Errorf("failed to create mail directory: %w", err)
	}

	pidPath := PIDFilePath(repoRoot)
	content := fmt.Sprintf("%d\n", pid)

	if err := os.WriteFile(pidPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// DeletePID removes the mailman.pid file.
// No error is returned if the file doesn't exist.
func DeletePID(repoRoot string) error {
	pidPath := PIDFilePath(repoRoot)

	if err := os.Remove(pidPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete PID file: %w", err)
	}

	return nil
}

// IsRunning checks if a process with the given PID is running.
// Returns false for PID 0 or if the process doesn't exist.
func IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// On Unix systems, sending signal 0 to a process checks if it exists
	// without actually sending a signal.
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Signal 0 doesn't send anything, just checks if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// StartDaemon starts the mailman daemon.
// If daemonize is false, runs in foreground mode and writes PID file.
// If daemonize is true, forks to background and parent exits immediately.
// Returns exit code: 0 on success, 2 if daemon already running.
func StartDaemon(repoRoot string, daemonize bool, stdout, stderr io.Writer) int {
	// Check if daemon is already running
	existingPID, err := ReadPID(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to read PID file: %v\n", err)
		return 1
	}

	if existingPID > 0 && IsRunning(existingPID) {
		fmt.Fprintf(stderr, "error: mailman daemon already running (PID: %d)\n", existingPID)
		return 2
	}

	// If we got here, either no PID file or stale PID - we can start

	if daemonize {
		// Background mode: fork and let parent exit
		return startBackground(repoRoot, stdout, stderr)
	}

	// Foreground mode: write PID and run directly
	return runForeground(repoRoot, stdout, stderr)
}

// runForeground runs the daemon in foreground mode.
// Writes PID file and outputs startup message.
func runForeground(repoRoot string, stdout, stderr io.Writer) int {
	currentPID := os.Getpid()

	// Write PID file
	if err := WritePID(repoRoot, currentPID); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	// Output startup message
	fmt.Fprintf(stdout, "Mailman daemon started (PID: %d)\n", currentPID)

	// In Phase 3, we just write PID and return.
	// The actual event loop will be added in Phase 6.
	return 0
}

// startBackground forks the daemon to background.
// Parent process outputs message and exits, child continues.
func startBackground(repoRoot string, stdout, stderr io.Writer) int {
	// Get the path to our own executable
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to get executable path: %v\n", err)
		return 1
	}

	// Start a new process with a special internal flag
	// The child will detect this flag and run in foreground mode
	cmd := &os.ProcAttr{
		Dir: repoRoot,
		Env: append(os.Environ(), "AGENTMAIL_DAEMON_CHILD=1"),
		Files: []*os.File{
			nil, // stdin - no input
			nil, // stdout - detached
			nil, // stderr - detached
		},
	}

	// Arguments: run mailman without --daemon (child will run foreground)
	args := []string{executable, "mailman"}

	process, err := os.StartProcess(executable, args, cmd)
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to start daemon process: %v\n", err)
		return 1
	}

	// Release the process so it's not a zombie
	if err := process.Release(); err != nil {
		fmt.Fprintf(stderr, "error: failed to release daemon process: %v\n", err)
		return 1
	}

	// Parent outputs background startup message
	fmt.Fprintf(stdout, "Mailman daemon started in background (PID: %d)\n", process.Pid)

	return 0
}

// IsDaemonChild returns true if this process was spawned as a daemon child.
// Used to detect when running as the background daemon process.
func IsDaemonChild() bool {
	return os.Getenv("AGENTMAIL_DAEMON_CHILD") == "1"
}
