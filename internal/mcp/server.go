package mcp

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"agentmail/internal/tmux"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version is set by build flags, defaults to "dev" for development builds.
var Version = "dev"

// Server wraps the MCP SDK server with AgentMail-specific functionality.
// It handles STDIO transport, tmux context validation, and tool registration.
type Server struct {
	mcpServer *mcp.Server
	logger    *log.Logger
}

// ServerOptions configures the MCP server behavior.
type ServerOptions struct {
	// SkipTmuxCheck disables tmux validation (for testing).
	SkipTmuxCheck bool
	// TmuxChecker is used to check tmux status (defaults to tmux.InTmux).
	TmuxChecker func() bool
}

// NewServer creates a new AgentMail MCP server.
// Returns an error if not running inside a tmux session (unless opts.SkipTmuxCheck is true).
func NewServer(opts *ServerOptions) (*Server, error) {
	if opts == nil {
		opts = &ServerOptions{}
	}

	// FR-015: Log errors and warnings to stderr
	logger := log.New(os.Stderr, "[agentmail-mcp] ", log.LstdFlags)

	// Check tmux context unless skipped (for testing)
	tmuxChecker := opts.TmuxChecker
	if tmuxChecker == nil {
		tmuxChecker = tmux.InTmux
	}

	if !opts.SkipTmuxCheck {
		if !tmuxChecker() {
			logger.Println("error: not running inside a tmux session")
			return nil, fmt.Errorf("not running inside a tmux session")
		}
	}

	// Create MCP server with implementation info
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "agentmail",
		Version: Version,
	}, nil)

	s := &Server{
		mcpServer: mcpServer,
		logger:    logger,
	}

	return s, nil
}

// Run starts the MCP server using STDIO transport.
// FR-001: Uses STDIO transport for all client communications.
// FR-010: Malformed JSON handling with -32700 error code is handled automatically
//
//	by the MCP SDK's underlying JSON-RPC library (jsonrpc.CodeParseError).
//
// FR-014: Monitors tmux context and exits within 1 second if lost.
func (s *Server) Run(ctx context.Context, opts *ServerOptions) error {
	if opts == nil {
		opts = &ServerOptions{}
	}

	// Determine tmux checker function
	tmuxChecker := opts.TmuxChecker
	if tmuxChecker == nil {
		tmuxChecker = tmux.InTmux
	}

	// FR-014: Create a context that cancels when tmux context is lost
	runCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	// Start tmux context monitoring goroutine
	if !opts.SkipTmuxCheck {
		go s.monitorTmuxContext(runCtx, cancel, tmuxChecker)
	}

	s.logger.Println("starting MCP server on STDIO transport")

	// FR-001: Run with STDIO transport
	err := s.mcpServer.Run(runCtx, &mcp.StdioTransport{})
	if err != nil {
		// Check if the error was due to tmux context loss
		if cause := context.Cause(runCtx); cause != nil && cause != context.Canceled {
			s.logger.Printf("error: %v", cause)
			return cause
		}
		// Don't log context.Canceled as an error - it's normal shutdown
		if err != context.Canceled {
			s.logger.Printf("error: server stopped: %v", err)
		}
		return err
	}

	return nil
}

// monitorTmuxContext periodically checks if tmux context is still available.
// FR-014: If tmux session terminates, exits within 1 second.
func (s *Server) monitorTmuxContext(ctx context.Context, cancel context.CancelCauseFunc, tmuxChecker func() bool) {
	// Check every 500ms to ensure we exit within 1 second of tmux loss
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !tmuxChecker() {
				s.logger.Println("tmux session terminated, shutting down")
				cancel(fmt.Errorf("tmux session terminated"))
				return
			}
		}
	}
}

// MCPServer returns the underlying MCP SDK server for tool registration.
// This allows the caller to register tools using mcp.AddTool().
func (s *Server) MCPServer() *mcp.Server {
	return s.mcpServer
}

// Logger returns the server's logger for consistent logging.
func (s *Server) Logger() *log.Logger {
	return s.logger
}
