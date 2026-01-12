package tmux

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// ErrNotInTmux is returned when tmux operations are attempted outside a tmux session.
var ErrNotInTmux = errors.New("not running inside a tmux session")

// ErrNoPaneID is returned when TMUX_PANE environment variable is not set.
var ErrNoPaneID = errors.New("TMUX_PANE environment variable not set")

// InTmux checks if the current process is running inside a tmux session.
// T010: Check $TMUX env var
func InTmux() bool {
	tmuxEnv := os.Getenv("TMUX")
	return tmuxEnv != ""
}

// GetCurrentPaneID returns the tmux pane ID where this process is running.
// The pane ID (e.g., "%0", "%3") is set by tmux at process spawn time and
// never changes, making it reliable for targeting even when windows are switched.
func GetCurrentPaneID() (string, error) {
	if !InTmux() {
		return "", ErrNotInTmux
	}

	paneID := os.Getenv("TMUX_PANE")
	if paneID == "" {
		return "", ErrNoPaneID
	}

	return paneID, nil
}

// GetCurrentWindow returns the name of the current tmux window.
// Uses TMUX_PANE to target the correct pane, avoiding race conditions
// when the user switches windows during command execution.
// T011: tmux display-message -t $TMUX_PANE -p '#W'
func GetCurrentWindow() (string, error) {
	paneID, err := GetCurrentPaneID()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("tmux", "display-message", "-t", paneID, "-p", "#W")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// ListWindows returns a list of all tmux window names in the current session.
// T012: tmux list-windows -F '#{window_name}'
func ListWindows() ([]string, error) {
	if !InTmux() {
		return nil, ErrNotInTmux
	}

	cmd := exec.Command("tmux", "list-windows", "-F", "#{window_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Split output into lines, filtering empty lines
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var windows []string
	for _, line := range lines {
		if line != "" {
			windows = append(windows, line)
		}
	}

	return windows, nil
}

// WindowExists checks if a window with the given name exists in the current tmux session.
// T013: Helper function
func WindowExists(name string) (bool, error) {
	windows, err := ListWindows()
	if err != nil {
		return false, err
	}

	for _, w := range windows {
		if w == name {
			return true, nil
		}
	}

	return false, nil
}
