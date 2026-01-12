package tmux

import (
	"testing"
)

// Tests for SendKeys and SendEnter functions

func TestSendKeys_NotInTmux(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	err := SendKeys("agent-1", "test message")
	if err == nil {
		t.Error("SendKeys() should return error when not in tmux")
	}
	if err != ErrNotInTmux {
		t.Errorf("SendKeys() should return ErrNotInTmux, got: %v", err)
	}
}

func TestSendEnter_NotInTmux(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	err := SendEnter("agent-1")
	if err == nil {
		t.Error("SendEnter() should return error when not in tmux")
	}
	if err != ErrNotInTmux {
		t.Errorf("SendEnter() should return ErrNotInTmux, got: %v", err)
	}
}
