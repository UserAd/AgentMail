package tmux

import (
	"testing"
)

// T004: Tests for tmux detection (InTmux, GetCurrentWindow)

func TestInTmux_WhenTMUXEnvSet(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "/tmp/tmux-501/default,12345,0")

	if !InTmux() {
		t.Error("InTmux() should return true when TMUX env var is set")
	}
}

func TestInTmux_WhenTMUXEnvNotSet(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	if InTmux() {
		t.Error("InTmux() should return false when TMUX env var is not set")
	}
}

func TestInTmux_WhenTMUXEnvEmpty(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	if InTmux() {
		t.Error("InTmux() should return false when TMUX env var is empty")
	}
}

// Tests for GetCurrentPaneID

func TestGetCurrentPaneID_NotInTmux(t *testing.T) {
	t.Setenv("TMUX", "")
	t.Setenv("TMUX_PANE", "%0")

	_, err := GetCurrentPaneID()
	if err != ErrNotInTmux {
		t.Errorf("GetCurrentPaneID() should return ErrNotInTmux when not in tmux, got: %v", err)
	}
}

func TestGetCurrentPaneID_NoPaneID(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-501/default,12345,0")
	t.Setenv("TMUX_PANE", "")

	_, err := GetCurrentPaneID()
	if err != ErrNoPaneID {
		t.Errorf("GetCurrentPaneID() should return ErrNoPaneID when TMUX_PANE is empty, got: %v", err)
	}
}

func TestGetCurrentPaneID_Success(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-501/default,12345,0")
	t.Setenv("TMUX_PANE", "%3")

	paneID, err := GetCurrentPaneID()
	if err != nil {
		t.Errorf("GetCurrentPaneID() should not return error, got: %v", err)
	}
	if paneID != "%3" {
		t.Errorf("GetCurrentPaneID() should return '%%3', got: %s", paneID)
	}
}

// T005: Tests for tmux window listing (ListWindows, WindowExists)

func TestListWindows_NotInTmux(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	_, err := ListWindows()
	if err == nil {
		t.Error("ListWindows() should return error when not in tmux")
	}
}

func TestWindowExists_NotInTmux(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	_, err := WindowExists("agent-1")
	if err == nil {
		t.Error("WindowExists() should return error when not in tmux")
	}
}

func TestGetCurrentWindow_NotInTmux(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")
	t.Setenv("TMUX_PANE", "")

	_, err := GetCurrentWindow()
	if err != ErrNotInTmux {
		t.Errorf("GetCurrentWindow() should return ErrNotInTmux when not in tmux, got: %v", err)
	}
}

func TestGetCurrentWindow_NoPaneID(t *testing.T) {
	// In tmux but TMUX_PANE not set (unusual but possible)
	t.Setenv("TMUX", "/tmp/tmux-501/default,12345,0")
	t.Setenv("TMUX_PANE", "")

	_, err := GetCurrentWindow()
	if err != ErrNoPaneID {
		t.Errorf("GetCurrentWindow() should return ErrNoPaneID when TMUX_PANE is empty, got: %v", err)
	}
}
