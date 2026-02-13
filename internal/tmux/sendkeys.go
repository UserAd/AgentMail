package tmux

import (
	"os/exec"
)

// SendKeys sends text to the specified tmux target.
// It executes: tmux send-keys -t <target> "<text>"
func SendKeys(target, text string) error {
	if !InTmux() {
		return ErrNotInTmux
	}

	cmd := exec.Command("tmux", "send-keys", "-t", target, text)
	return cmd.Run()
}

// SendEnter sends an Enter keypress to the specified tmux target.
// It executes: tmux send-keys -t <target> Enter
func SendEnter(target string) error {
	if !InTmux() {
		return ErrNotInTmux
	}

	cmd := exec.Command("tmux", "send-keys", "-t", target, "Enter")
	return cmd.Run()
}
