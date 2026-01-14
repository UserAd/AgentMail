package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupTestMailbox creates a temporary directory with the .agentmail structure
// and returns the path. Caller is responsible for cleanup.
func setupTestMailbox(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "agentmail-mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	return tmpDir
}

// writeTestMessages writes test messages to a recipient's mailbox file.
func writeTestMessages(t *testing.T, repoRoot, recipient, content string) string {
	t.Helper()

	mailDir := filepath.Join(repoRoot, ".agentmail", "mailboxes")
	filePath := filepath.Join(mailDir, recipient+".jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	return filePath
}

// T019: Test receive handler returns oldest unread message (FR-003)
func TestReceiveHandler_ReturnsOldestUnreadMessage(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create mailbox with multiple unread messages
	content := `{"id":"first123","from":"agent-1","to":"agent-2","message":"First message","read_flag":false}
{"id":"second12","from":"agent-3","to":"agent-2","message":"Second message","read_flag":false}
{"id":"third123","from":"agent-1","to":"agent-2","message":"Third message","read_flag":false}
`
	writeTestMessages(t, tmpDir, "agent-2", content)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("receiveHandler returned error result: %v", result.Content)
	}

	// Parse the response
	if len(result.Content) == 0 {
		t.Fatal("receiveHandler returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("receiveHandler content is not TextContent, got %T", result.Content[0])
	}

	var response ReceiveResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// FR-003: Should return the oldest (first) message
	if response.ID != "first123" {
		t.Errorf("Expected ID 'first123' (oldest), got '%s'", response.ID)
	}
	if response.From != "agent-1" {
		t.Errorf("Expected From 'agent-1', got '%s'", response.From)
	}
	if response.Message != "First message" {
		t.Errorf("Expected Message 'First message', got '%s'", response.Message)
	}
}

// T020: Test receive with no messages returns "No unread messages" (FR-008)
func TestReceiveHandler_NoMessagesReturnsEmptyStatus(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// No messages written - empty mailbox

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("receiveHandler returned error result: %v", result.Content)
	}

	// Parse the response
	if len(result.Content) == 0 {
		t.Fatal("receiveHandler returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("receiveHandler content is not TextContent, got %T", result.Content[0])
	}

	var response ReceiveEmptyResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// FR-008: Should return status "No unread messages"
	if response.Status != "No unread messages" {
		t.Errorf("Expected status 'No unread messages', got '%s'", response.Status)
	}
}

// Test receive with all messages already read returns "No unread messages"
func TestReceiveHandler_AllMessagesReadReturnsEmptyStatus(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// All messages marked as read
	content := `{"id":"id1","from":"agent-1","to":"agent-2","message":"Hello","read_flag":true}
{"id":"id2","from":"agent-3","to":"agent-2","message":"World","read_flag":true}
`
	writeTestMessages(t, tmpDir, "agent-2", content)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("receiveHandler returned error result: %v", result.Content)
	}

	// Parse the response
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("receiveHandler content is not TextContent")
	}

	var response ReceiveEmptyResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should return status "No unread messages"
	if response.Status != "No unread messages" {
		t.Errorf("Expected status 'No unread messages', got '%s'", response.Status)
	}
}

// T021: Test message is marked as read after receive (FR-012)
func TestReceiveHandler_MessageMarkedAsReadAfterReceive(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create mailbox with one unread message
	content := `{"id":"testID01","from":"agent-1","to":"agent-2","message":"Test message","read_flag":false}
`
	filePath := writeTestMessages(t, tmpDir, "agent-2", content)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("receiveHandler returned error result: %v", result.Content)
	}

	// FR-012: Verify message was marked as read in the file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read mailbox file: %v", err)
	}

	if !strings.Contains(string(fileContent), `"read_flag":true`) {
		t.Errorf("Message should be marked as read after receive. File content: %s", fileContent)
	}
}

// Test receive returns correct JSON fields per data-model.md
func TestReceiveHandler_ResponseFieldsMatchDataModel(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create mailbox with a message
	content := `{"id":"abc12345","from":"sender-agent","to":"receiver-agent","message":"Hello, World!","read_flag":false}
`
	writeTestMessages(t, tmpDir, "receiver-agent", content)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "receiver-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)

	// Verify JSON structure matches data-model.md
	var responseMap map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &responseMap); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Check required fields exist
	requiredFields := []string{"from", "id", "message"}
	for _, field := range requiredFields {
		if _, ok := responseMap[field]; !ok {
			t.Errorf("Response missing required field '%s'", field)
		}
	}

	// Verify field values
	if responseMap["from"] != "sender-agent" {
		t.Errorf("Expected from='sender-agent', got '%v'", responseMap["from"])
	}
	if responseMap["id"] != "abc12345" {
		t.Errorf("Expected id='abc12345', got '%v'", responseMap["id"])
	}
	if responseMap["message"] != "Hello, World!" {
		t.Errorf("Expected message='Hello, World!', got '%v'", responseMap["message"])
	}
}

// T026 / SC-003: Verify MCP receive output semantic equivalence with CLI output
func TestReceiveHandler_OutputMatchesCLIFormat(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create mailbox with a message
	content := `{"id":"testID99","from":"cli-sender","to":"cli-receiver","message":"CLI comparison test","read_flag":false}
`
	writeTestMessages(t, tmpDir, "cli-receiver", content)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "cli-receiver",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)

	var response ReceiveResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// CLI output format is:
	// From: <sender>
	// ID: <id>
	//
	// <message>
	//
	// MCP output should contain the same information in JSON format:
	// {"from": "<sender>", "id": "<id>", "message": "<message>"}
	//
	// SC-003: Verify semantic equivalence - same data, different format
	if response.From != "cli-sender" {
		t.Errorf("From field mismatch: expected 'cli-sender', got '%s'", response.From)
	}
	if response.ID != "testID99" {
		t.Errorf("ID field mismatch: expected 'testID99', got '%s'", response.ID)
	}
	if response.Message != "CLI comparison test" {
		t.Errorf("Message field mismatch: expected 'CLI comparison test', got '%s'", response.Message)
	}
}

// Test consecutive receives return messages in FIFO order
func TestReceiveHandler_ConsecutiveReceivesFIFOOrder(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create mailbox with three unread messages
	content := `{"id":"msg001","from":"agent-a","to":"agent-b","message":"First","read_flag":false}
{"id":"msg002","from":"agent-c","to":"agent-b","message":"Second","read_flag":false}
{"id":"msg003","from":"agent-a","to":"agent-b","message":"Third","read_flag":false}
`
	writeTestMessages(t, tmpDir, "agent-b", content)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-b",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	ctx := context.Background()

	// First receive should return msg001
	result1, _ := receiveHandler(ctx, &mcp.CallToolRequest{})
	text1 := result1.Content[0].(*mcp.TextContent).Text
	var resp1 ReceiveResponse
	json.Unmarshal([]byte(text1), &resp1)
	if resp1.ID != "msg001" {
		t.Errorf("First receive: expected 'msg001', got '%s'", resp1.ID)
	}

	// Second receive should return msg002
	result2, _ := receiveHandler(ctx, &mcp.CallToolRequest{})
	text2 := result2.Content[0].(*mcp.TextContent).Text
	var resp2 ReceiveResponse
	json.Unmarshal([]byte(text2), &resp2)
	if resp2.ID != "msg002" {
		t.Errorf("Second receive: expected 'msg002', got '%s'", resp2.ID)
	}

	// Third receive should return msg003
	result3, _ := receiveHandler(ctx, &mcp.CallToolRequest{})
	text3 := result3.Content[0].(*mcp.TextContent).Text
	var resp3 ReceiveResponse
	json.Unmarshal([]byte(text3), &resp3)
	if resp3.ID != "msg003" {
		t.Errorf("Third receive: expected 'msg003', got '%s'", resp3.ID)
	}

	// Fourth receive should return empty status
	result4, _ := receiveHandler(ctx, &mcp.CallToolRequest{})
	text4 := result4.Content[0].(*mcp.TextContent).Text
	var resp4 ReceiveEmptyResponse
	json.Unmarshal([]byte(text4), &resp4)
	if resp4.Status != "No unread messages" {
		t.Errorf("Fourth receive: expected 'No unread messages', got '%s'", resp4.Status)
	}
}

// Test receive skips already-read messages
func TestReceiveHandler_SkipsReadMessages(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Mix of read and unread messages
	content := `{"id":"read001","from":"agent-1","to":"agent-2","message":"Already read","read_flag":true}
{"id":"unread01","from":"agent-1","to":"agent-2","message":"Not read yet","read_flag":false}
{"id":"read002","from":"agent-1","to":"agent-2","message":"Also read","read_flag":true}
`
	writeTestMessages(t, tmpDir, "agent-2", content)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)

	var response ReceiveResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should return the unread message, not the read ones
	if response.ID != "unread01" {
		t.Errorf("Expected unread message 'unread01', got '%s'", response.ID)
	}
}

// Test empty response JSON structure matches data-model.md
func TestReceiveHandler_EmptyResponseStructure(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing (empty mailbox)
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)

	// Verify JSON structure matches data-model.md ReceiveEmptyResponse
	var responseMap map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &responseMap); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should have only "status" field
	if len(responseMap) != 1 {
		t.Errorf("Expected 1 field, got %d: %v", len(responseMap), responseMap)
	}
	if _, ok := responseMap["status"]; !ok {
		t.Error("Response missing 'status' field")
	}
}

// Integration test via MCP client
func TestReceiveHandler_MCPClientIntegration(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create mailbox with a message
	content := `{"id":"mcptest1","from":"mcp-sender","to":"mcp-receiver","message":"MCP integration test","read_flag":false}
`
	writeTestMessages(t, tmpDir, "mcp-receiver", content)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "mcp-receiver",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Set up test server and client
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()

	// Call receive tool via MCP client
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolReceive,
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool(receive) failed: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("CallTool(receive) returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("CallTool(receive) content is not TextContent")
	}

	// Parse and verify response
	var response ReceiveResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response.ID != "mcptest1" {
		t.Errorf("Expected ID 'mcptest1', got '%s'", response.ID)
	}
	if response.From != "mcp-sender" {
		t.Errorf("Expected From 'mcp-sender', got '%s'", response.From)
	}
	if response.Message != "MCP integration test" {
		t.Errorf("Expected Message 'MCP integration test', got '%s'", response.Message)
	}
}
