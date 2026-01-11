package tmux

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// ErrNotInTmux is returned when tmux operations are attempted outside a tmux session.
var ErrNotInTmux = errors.New("not running inside a tmux session")

// InTmux checks if the current process is running inside a tmux session.
// T010: Check $TMUX env var
func InTmux() bool {
	tmuxEnv := os.Getenv("TMUX")
	return tmuxEnv != ""
}

// GetCurrentWindow returns the name of the current tmux window.
// T011: tmux display-message -p '#W'
func GetCurrentWindow() (string, error) {
	if !InTmux() {
		return "", ErrNotInTmux
	}

	cmd := exec.Command("tmux", "display-message", "-p", "#W")
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
