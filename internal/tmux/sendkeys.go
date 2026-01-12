package tmux

import (
	"os/exec"
)

// SendKeys sends text to the specified tmux window.
// It executes: tmux send-keys -t <window> "<text>"
func SendKeys(window, text string) error {
	if !InTmux() {
		return ErrNotInTmux
	}

	cmd := exec.Command("tmux", "send-keys", "-t", window, text)
	return cmd.Run()
}

// SendEnter sends an Enter keypress to the specified tmux window.
// It executes: tmux send-keys -t <window> Enter
func SendEnter(window string) error {
	if !InTmux() {
		return ErrNotInTmux
	}

	cmd := exec.Command("tmux", "send-keys", "-t", window, "Enter")
	return cmd.Run()
}
