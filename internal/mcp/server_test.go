package mcp

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewServer_NotInTmux(t *testing.T) {
	// Test that NewServer returns error when not in tmux
	_, err := NewServer(&ServerOptions{
		TmuxChecker: func() bool { return false },
	})

	if err == nil {
		t.Error("NewServer should return error when not in tmux")
	}
	if err.Error() != "not running inside a tmux session" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewServer_InTmux(t *testing.T) {
	// Test that NewServer succeeds when in tmux
	server, err := NewServer(&ServerOptions{
		TmuxChecker: func() bool { return true },
	})

	if err != nil {
		t.Errorf("NewServer should not return error when in tmux: %v", err)
	}
	if server == nil {
		t.Error("NewServer should return non-nil server")
	}
	if server.mcpServer == nil {
		t.Error("NewServer should create underlying MCP server")
	}
	if server.logger == nil {
		t.Error("NewServer should create logger")
	}
}

func TestNewServer_SkipTmuxCheck(t *testing.T) {
	// Test that SkipTmuxCheck bypasses tmux validation
	server, err := NewServer(&ServerOptions{
		SkipTmuxCheck: true,
		TmuxChecker:   func() bool { return false }, // Would fail without skip
	})

	if err != nil {
		t.Errorf("NewServer should not return error with SkipTmuxCheck: %v", err)
	}
	if server == nil {
		t.Error("NewServer should return non-nil server with SkipTmuxCheck")
	}
}

func TestNewServer_NilOptions(t *testing.T) {
	// Test that nil options uses real tmux check
	// We simulate non-tmux environment to verify it returns an error
	t.Setenv("TMUX", "")

	_, err := NewServer(nil)
	if err == nil {
		t.Error("NewServer with nil options should fail outside tmux")
	}
}

func TestServer_MCPServer(t *testing.T) {
	// Test that MCPServer returns the underlying MCP server
	server, err := NewServer(&ServerOptions{
		SkipTmuxCheck: true,
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	mcpServer := server.MCPServer()
	if mcpServer == nil {
		t.Error("MCPServer should return non-nil MCP server")
	}
}

func TestServer_Logger(t *testing.T) {
	// Test that Logger returns the server's logger
	server, err := NewServer(&ServerOptions{
		SkipTmuxCheck: true,
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	logger := server.Logger()
	if logger == nil {
		t.Error("Logger should return non-nil logger")
	}
}

func TestServer_MonitorTmuxContext(t *testing.T) {
	// Test that tmux context monitoring detects loss
	server, err := NewServer(&ServerOptions{
		SkipTmuxCheck: true,
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// Use atomic for thread-safe access to tmux state
	var inTmux atomic.Bool
	inTmux.Store(true)

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	// Start monitoring
	go server.monitorTmuxContext(ctx, cancel, func() bool {
		return inTmux.Load()
	})

	// Simulate tmux loss
	inTmux.Store(false)

	// Wait for monitoring to detect the loss (should be within 1 second per FR-014)
	select {
	case <-ctx.Done():
		// Expected - context should be canceled
		cause := context.Cause(ctx)
		if cause == nil {
			t.Error("context should have a cause set")
		} else if cause.Error() != "tmux session terminated" {
			t.Errorf("unexpected cause: %v", cause)
		}
	case <-time.After(1100 * time.Millisecond):
		t.Error("tmux context loss should be detected within 1 second")
	}
}

func TestServer_MonitorTmuxContext_ContextCanceled(t *testing.T) {
	// Test that monitoring stops when context is canceled
	server, err := NewServer(&ServerOptions{
		SkipTmuxCheck: true,
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	done := make(chan struct{})

	// Start monitoring
	go func() {
		server.monitorTmuxContext(ctx, cancel, func() bool { return true })
		close(done)
	}()

	// Cancel context
	cancel(nil)

	// Monitor should exit
	select {
	case <-done:
		// Expected
	case <-time.After(600 * time.Millisecond):
		t.Error("monitoring should stop when context is canceled")
	}
}

func TestServer_Run_SkipTmuxCheck(t *testing.T) {
	// Test that Run with SkipTmuxCheck starts without tmux monitoring
	server, err := NewServer(&ServerOptions{
		SkipTmuxCheck: true,
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// Create a context that we'll cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Run should exit due to context cancellation
	err = server.Run(ctx, &ServerOptions{
		SkipTmuxCheck: true,
	})

	// We expect context.Canceled since we canceled immediately
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("Run should return context.Canceled or nil, got: %v", err)
	}
}

func TestVersion_Default(t *testing.T) {
	// Test that Version has a default value
	if Version == "" {
		t.Error("Version should have a default value")
	}
	if Version != "dev" {
		t.Logf("Version is set to: %s (may be set by build flags)", Version)
	}
}

func TestVersion_CanBeOverridden(t *testing.T) {
	// This test documents how to override the Version constant at build time.
	// The Version variable can be set via ldflags during compilation:
	//
	//   go build -ldflags="-X agentmail/internal/mcp.Version=1.0.0" ./cmd/agentmail
	//
	// For releases, the GitHub Actions workflow sets this automatically:
	//
	//   go build -ldflags="-s -w -X main.version=${VERSION}" ./cmd/agentmail
	//
	// Note: The version is typically set on main.version, which is then used
	// to initialize mcp.Version or passed to the MCP server.
	//
	// This test cannot verify ldflags behavior at runtime, but documents the pattern.
	t.Log("Version can be overridden at build time using:")
	t.Log("  go build -ldflags=\"-X agentmail/internal/mcp.Version=1.0.0\" ./cmd/agentmail")
	t.Logf("Current Version value: %s", Version)
}
