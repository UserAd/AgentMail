package tmux

import (
	"os"
	"testing"
)

// T004: Tests for tmux detection (InTmux, GetCurrentWindow)

func TestInTmux_WhenTMUXEnvSet(t *testing.T) {
	// Save original value
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)

	// Set TMUX env var
	os.Setenv("TMUX", "/tmp/tmux-501/default,12345,0")

	if !InTmux() {
		t.Error("InTmux() should return true when TMUX env var is set")
	}
}

func TestInTmux_WhenTMUXEnvNotSet(t *testing.T) {
	// Save original value
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)

	// Unset TMUX env var
	os.Unsetenv("TMUX")

	if InTmux() {
		t.Error("InTmux() should return false when TMUX env var is not set")
	}
}

func TestInTmux_WhenTMUXEnvEmpty(t *testing.T) {
	// Save original value
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)

	// Set TMUX to empty string
	os.Setenv("TMUX", "")

	if InTmux() {
		t.Error("InTmux() should return false when TMUX env var is empty")
	}
}

// T005: Tests for tmux window listing (ListWindows, WindowExists)

func TestListWindows_NotInTmux(t *testing.T) {
	// Save original value
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)

	// Unset TMUX env var
	os.Unsetenv("TMUX")

	_, err := ListWindows()
	if err == nil {
		t.Error("ListWindows() should return error when not in tmux")
	}
}

func TestWindowExists_NotInTmux(t *testing.T) {
	// Save original value
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)

	// Unset TMUX env var
	os.Unsetenv("TMUX")

	_, err := WindowExists("agent-1")
	if err == nil {
		t.Error("WindowExists() should return error when not in tmux")
	}
}

func TestGetCurrentWindow_NotInTmux(t *testing.T) {
	// Save original value
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)

	// Unset TMUX env var
	os.Unsetenv("TMUX")

	_, err := GetCurrentWindow()
	if err == nil {
		t.Error("GetCurrentWindow() should return error when not in tmux")
	}
}
