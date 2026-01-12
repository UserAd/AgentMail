// Package daemon provides functionality for the mailman daemon process.
// It handles PID file management and daemon startup/shutdown.
package daemon

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"agentmail/internal/mail"
)

// PIDFile is the filename for the mailman daemon PID file within .git/mail/
const PIDFile = "mailman.pid"

// DaemonStatus represents the status of an existing daemon process.
type DaemonStatus int

const (
	// DaemonNone indicates no daemon is running and no PID file exists.
	DaemonNone DaemonStatus = iota
	// DaemonRunning indicates a daemon is currently running.
	DaemonRunning
	// DaemonStale indicates a PID file exists but the process is not running.
	DaemonStale
)

// stopChan is used for testing to allow stopping the daemon without signals.
// When this is non-nil, the daemon will select on both signals and this channel.
var stopChan chan struct{}

// SetStopChannel sets a channel that can be used to stop the daemon for testing.
// Pass nil to disable test mode. This is only for testing purposes.
func SetStopChannel(ch chan struct{}) {
	stopChan = ch
}

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

// CheckExistingDaemon checks if a daemon is already running.
// Returns:
//   - DaemonStatus: Running, Stale, or None
//   - int: PID (if found, otherwise 0)
//   - error: if file read/parse fails
func CheckExistingDaemon(repoRoot string) (DaemonStatus, int, error) {
	pid, err := ReadPID(repoRoot)
	if err != nil {
		return DaemonNone, 0, err
	}

	if pid == 0 {
		// No PID file exists
		return DaemonNone, 0, nil
	}

	if IsRunning(pid) {
		return DaemonRunning, pid, nil
	}

	// PID file exists but process is not running - stale
	return DaemonStale, pid, nil
}

// StartDaemon starts the mailman daemon.
// If daemonize is false, runs in foreground mode and writes PID file.
// If daemonize is true, forks to background and parent exits immediately.
// Returns exit code: 0 on success, 2 if daemon already running.
func StartDaemon(repoRoot string, daemonize bool, stdout, stderr io.Writer) int {
	// Check if daemon is already running
	status, pid, err := CheckExistingDaemon(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to read PID file: %v\n", err)
		return 1
	}

	switch status {
	case DaemonRunning:
		fmt.Fprintf(stderr, "error: mailman daemon already running (PID: %d)\n", pid)
		return 2
	case DaemonStale:
		// Clean up stale PID file with warning
		fmt.Fprintf(stderr, "Warning: Stale PID file found, cleaning up\n")
		if err := DeletePID(repoRoot); err != nil {
			fmt.Fprintf(stderr, "error: failed to clean up stale PID file: %v\n", err)
			return 1
		}
	case DaemonNone:
		// No existing daemon, proceed with startup
	}

	if daemonize {
		// Background mode: fork and let parent exit
		return startBackground(repoRoot, stdout, stderr)
	}

	// Foreground mode: write PID and run directly
	return runForeground(repoRoot, stdout, stderr)
}

// runForeground runs the daemon in foreground mode.
// Writes PID file and outputs startup message.
// Sets up signal handling for graceful shutdown on SIGTERM/SIGINT.
func runForeground(repoRoot string, stdout, stderr io.Writer) int {
	currentPID := os.Getpid()

	// Write PID file
	if err := WritePID(repoRoot, currentPID); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Output startup message
	fmt.Fprintf(stdout, "Mailman daemon started (PID: %d)\n", currentPID)

	// Wait for shutdown signal or test stop
	// In Phase 3, we wait for signal and then clean up.
	// The actual event loop will be added in Phase 6.
	if stopChan != nil {
		// Test mode: wait on either signal or stop channel
		select {
		case <-sigChan:
		case <-stopChan:
		}
	} else {
		// Production mode: wait only on signals
		<-sigChan
	}

	// Clean up PID file on shutdown
	DeletePID(repoRoot)

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
