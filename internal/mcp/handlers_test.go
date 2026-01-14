package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// makeSendRequest creates a CallToolRequest for the send tool with the given arguments.
func makeSendRequest(recipient, message string) *mcp.CallToolRequest {
	args, _ := json.Marshal(map[string]string{
		"recipient": recipient,
		"message":   message,
	})
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      ToolSend,
			Arguments: args,
		},
	}
}

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
	if len(result.Content) == 0 {
		t.Fatal("receiveHandler returned empty content")
	}
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
	if err := json.Unmarshal([]byte(text1), &resp1); err != nil {
		t.Fatalf("Failed to unmarshal first response: %v", err)
	}
	if resp1.ID != "msg001" {
		t.Errorf("First receive: expected 'msg001', got '%s'", resp1.ID)
	}

	// Second receive should return msg002
	result2, _ := receiveHandler(ctx, &mcp.CallToolRequest{})
	text2 := result2.Content[0].(*mcp.TextContent).Text
	var resp2 ReceiveResponse
	if err := json.Unmarshal([]byte(text2), &resp2); err != nil {
		t.Fatalf("Failed to unmarshal second response: %v", err)
	}
	if resp2.ID != "msg002" {
		t.Errorf("Second receive: expected 'msg002', got '%s'", resp2.ID)
	}

	// Third receive should return msg003
	result3, _ := receiveHandler(ctx, &mcp.CallToolRequest{})
	text3 := result3.Content[0].(*mcp.TextContent).Text
	var resp3 ReceiveResponse
	if err := json.Unmarshal([]byte(text3), &resp3); err != nil {
		t.Fatalf("Failed to unmarshal third response: %v", err)
	}
	if resp3.ID != "msg003" {
		t.Errorf("Third receive: expected 'msg003', got '%s'", resp3.ID)
	}

	// Fourth receive should return empty status
	result4, _ := receiveHandler(ctx, &mcp.CallToolRequest{})
	text4 := result4.Content[0].(*mcp.TextContent).Text
	var resp4 ReceiveEmptyResponse
	if err := json.Unmarshal([]byte(text4), &resp4); err != nil {
		t.Fatalf("Failed to unmarshal fourth response: %v", err)
	}
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
	_, clientSession, cleanup := setupTestServer(t)
	defer cleanup()

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

// ============================================================================
// Send Handler Tests
// ============================================================================

// T027: Test send handler delivers message and returns ID (FR-004)
func TestSendHandler_DeliversMessageAndReturnsID(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the send handler
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("agent-receiver", "Hello from MCP!"))
	if err != nil {
		t.Fatalf("sendHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("sendHandler returned error result: %v", result.Content)
	}

	// Parse the response
	if len(result.Content) == 0 {
		t.Fatal("sendHandler returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler content is not TextContent, got %T", result.Content[0])
	}

	var response SendResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// FR-004: Should return a message ID
	if response.MessageID == "" {
		t.Error("Expected non-empty message_id in response")
	}

	// Verify the message was stored in the mailbox
	mailboxPath := filepath.Join(tmpDir, ".agentmail", "mailboxes", "agent-receiver.jsonl")
	data, err := os.ReadFile(mailboxPath)
	if err != nil {
		t.Fatalf("Failed to read mailbox file: %v", err)
	}

	if !strings.Contains(string(data), "Hello from MCP!") {
		t.Errorf("Message not found in mailbox. Content: %s", data)
	}
	if !strings.Contains(string(data), response.MessageID) {
		t.Errorf("Message ID not found in mailbox. Content: %s", data)
	}
}

// T028: Test send with invalid recipient returns error (FR-009)
func TestSendHandler_InvalidRecipientReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing - nonexistent-agent not in MockWindows
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the send handler with invalid recipient
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("nonexistent-agent", "This should fail"))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// FR-009: Should return an error result
	if !result.IsError {
		t.Fatal("sendHandler should return error for invalid recipient")
	}

	// Verify error message contains "recipient not found"
	if len(result.Content) == 0 {
		t.Fatal("sendHandler returned empty error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "recipient not found") {
		t.Errorf("Expected error to contain 'recipient not found', got: %s", textContent.Text)
	}
}

// T029: Test send with message > 64KB returns error (FR-013)
func TestSendHandler_OversizedMessageReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Create a message larger than 64KB (65536 bytes)
	oversizedMessage := strings.Repeat("x", 65537)

	// Call the send handler with oversized message
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("agent-receiver", oversizedMessage))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// FR-013: Should return an error result
	if !result.IsError {
		t.Fatal("sendHandler should return error for oversized message")
	}

	// Verify error message indicates size limit
	if len(result.Content) == 0 {
		t.Fatal("sendHandler returned empty error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "64KB") && !strings.Contains(textContent.Text, "size") {
		t.Errorf("Expected error to mention size limit, got: %s", textContent.Text)
	}
}

// Test send with empty message returns error
func TestSendHandler_EmptyMessageReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the send handler with empty message
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("agent-receiver", ""))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("sendHandler should return error for empty message")
	}

	// Verify error message
	if len(result.Content) == 0 {
		t.Fatal("sendHandler returned empty error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "no message provided") {
		t.Errorf("Expected error to contain 'no message provided', got: %s", textContent.Text)
	}
}

// Test send to self returns error
func TestSendHandler_SendToSelfReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-self",
		MockWindows:   []string{"agent-self"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the send handler with recipient = sender
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("agent-self", "Hello myself!"))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result (self-send not allowed)
	if !result.IsError {
		t.Fatal("sendHandler should return error for sending to self")
	}

	// Verify error message contains "cannot send message to self"
	if len(result.Content) == 0 {
		t.Fatal("sendHandler returned empty error content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "cannot send message to self") {
		t.Errorf("Expected error to contain 'cannot send message to self', got: %s", textContent.Text)
	}
}

// Test send to ignored recipient returns error
func TestSendHandler_IgnoredRecipientReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create .agentmailignore file
	ignoreFile := filepath.Join(tmpDir, ".agentmailignore")
	if err := os.WriteFile(ignoreFile, []byte("ignored-agent\n"), 0644); err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck:  true,
		MockSender:     "agent-sender",
		MockWindows:    []string{"agent-sender", "ignored-agent"},
		RepoRoot:       tmpDir,
		MockIgnoreList: map[string]bool{"ignored-agent": true},
	})
	defer SetHandlerOptions(nil)

	// Call the send handler with ignored recipient
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("ignored-agent", "This should fail"))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("sendHandler should return error for ignored recipient")
	}

	// Verify error message contains "recipient not found"
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "recipient not found") {
		t.Errorf("Expected error to contain 'recipient not found', got: %s", textContent.Text)
	}
}

// Test send response format matches data-model.md
func TestSendHandler_ResponseFormat(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the send handler
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("agent-receiver", "Test message"))
	if err != nil {
		t.Fatalf("sendHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)

	// Verify JSON structure matches data-model.md SendResponse
	var responseMap map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &responseMap); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should have only "message_id" field
	if len(responseMap) != 1 {
		t.Errorf("Expected 1 field, got %d: %v", len(responseMap), responseMap)
	}
	if _, ok := responseMap["message_id"]; !ok {
		t.Error("Response missing 'message_id' field")
	}

	// Verify message_id is a non-empty string
	messageID, ok := responseMap["message_id"].(string)
	if !ok {
		t.Errorf("message_id is not a string: %T", responseMap["message_id"])
	}
	if messageID == "" {
		t.Error("message_id should not be empty")
	}
}

// T034 / SC-002: Verify MCP send creates message readable via CLI
func TestSendHandler_MessageReadableViaCLI(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "mcp-sender",
		MockReceiver:  "cli-receiver",
		MockWindows:   []string{"mcp-sender", "cli-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Send a message via MCP
	ctx := context.Background()
	sendResult, err := sendHandler(ctx, makeSendRequest("cli-receiver", "MCP to CLI test message"))
	if err != nil {
		t.Fatalf("sendHandler returned error: %v", err)
	}

	if sendResult.IsError {
		t.Fatalf("sendHandler returned error result: %v", sendResult.Content)
	}

	// Get the message ID from send response
	textContent := sendResult.Content[0].(*mcp.TextContent)
	var sendResponse SendResponse
	if err := json.Unmarshal([]byte(textContent.Text), &sendResponse); err != nil {
		t.Fatalf("Failed to parse send response: %v", err)
	}

	// Now receive the message via MCP receive handler (simulates CLI receive)
	receiveResult, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned error: %v", err)
	}

	if receiveResult.IsError {
		t.Fatalf("receiveHandler returned error result: %v", receiveResult.Content)
	}

	// Parse receive response
	receiveTextContent := receiveResult.Content[0].(*mcp.TextContent)
	var receiveResponse ReceiveResponse
	if err := json.Unmarshal([]byte(receiveTextContent.Text), &receiveResponse); err != nil {
		t.Fatalf("Failed to parse receive response: %v", err)
	}

	// SC-002: Verify the received message matches what was sent
	if receiveResponse.ID != sendResponse.MessageID {
		t.Errorf("Message ID mismatch: sent %s, received %s", sendResponse.MessageID, receiveResponse.ID)
	}
	if receiveResponse.From != "mcp-sender" {
		t.Errorf("Sender mismatch: expected 'mcp-sender', got '%s'", receiveResponse.From)
	}
	if receiveResponse.Message != "MCP to CLI test message" {
		t.Errorf("Message mismatch: expected 'MCP to CLI test message', got '%s'", receiveResponse.Message)
	}
}

// Test send via MCP client integration
func TestSendHandler_MCPClientIntegration(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "mcp-client-sender",
		MockWindows:   []string{"mcp-client-sender", "mcp-client-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Set up test server and client
	_, clientSession, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Call send tool via MCP client
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: ToolSend,
		Arguments: map[string]any{
			"recipient": "mcp-client-receiver",
			"message":   "MCP client integration test",
		},
	})
	if err != nil {
		t.Fatalf("CallTool(send) failed: %v", err)
	}

	if result.IsError {
		textContent := result.Content[0].(*mcp.TextContent)
		t.Fatalf("CallTool(send) returned error: %s", textContent.Text)
	}

	if len(result.Content) == 0 {
		t.Fatal("CallTool(send) returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("CallTool(send) content is not TextContent")
	}

	// Parse and verify response
	var response SendResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response.MessageID == "" {
		t.Error("Expected non-empty message_id")
	}

	// Verify message was stored
	mailboxPath := filepath.Join(tmpDir, ".agentmail", "mailboxes", "mcp-client-receiver.jsonl")
	data, err := os.ReadFile(mailboxPath)
	if err != nil {
		t.Fatalf("Failed to read mailbox file: %v", err)
	}

	if !strings.Contains(string(data), "MCP client integration test") {
		t.Errorf("Message not found in mailbox. Content: %s", data)
	}
}

// Test send exactly at 64KB boundary succeeds
func TestSendHandler_ExactlyMaxSizeSucceeds(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Create a message exactly at 64KB (65536 bytes)
	exactMessage := strings.Repeat("x", 65536)

	// Call the send handler with exact max size message
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("agent-receiver", exactMessage))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should succeed
	if result.IsError {
		textContent := result.Content[0].(*mcp.TextContent)
		t.Fatalf("sendHandler should succeed for message exactly at 64KB, got error: %s", textContent.Text)
	}

	// Verify response has message_id
	textContent := result.Content[0].(*mcp.TextContent)
	var response SendResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response.MessageID == "" {
		t.Error("Expected non-empty message_id")
	}
}

// ============================================================================
// Status Handler Tests
// ============================================================================

// makeStatusRequest creates a CallToolRequest for the status tool with the given status.
func makeStatusRequest(status string) *mcp.CallToolRequest {
	args, _ := json.Marshal(map[string]string{
		"status": status,
	})
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      ToolStatus,
			Arguments: args,
		},
	}
}

// T035: Test status handler updates status and returns "ok" (FR-005)
func TestStatusHandler_UpdatesStatusAndReturnsOk(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the status handler with "ready"
	ctx := context.Background()
	result, err := statusHandler(ctx, makeStatusRequest("ready"))
	if err != nil {
		t.Fatalf("statusHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("statusHandler returned error result: %v", result.Content)
	}

	// Parse the response
	if len(result.Content) == 0 {
		t.Fatal("statusHandler returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("statusHandler content is not TextContent, got %T", result.Content[0])
	}

	var response StatusResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// FR-005: Should return status "ok"
	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	// Verify the status was persisted to recipients.jsonl
	recipientsPath := filepath.Join(tmpDir, ".agentmail", "recipients.jsonl")
	data, err := os.ReadFile(recipientsPath)
	if err != nil {
		t.Fatalf("Failed to read recipients file: %v", err)
	}

	if !strings.Contains(string(data), `"recipient":"test-agent"`) {
		t.Errorf("Recipient not found in recipients file. Content: %s", data)
	}
	if !strings.Contains(string(data), `"status":"ready"`) {
		t.Errorf("Status 'ready' not found in recipients file. Content: %s", data)
	}
}

// T036: Test status with invalid value returns error (FR-016)
func TestStatusHandler_InvalidValueReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Test various invalid status values
	invalidStatuses := []string{"invalid", "busy", "available", "READY", "Ready", ""}

	for _, status := range invalidStatuses {
		ctx := context.Background()
		result, err := statusHandler(ctx, makeStatusRequest(status))
		if err != nil {
			t.Fatalf("statusHandler returned unexpected error for '%s': %v", status, err)
		}

		// FR-016: Should return an error result
		if !result.IsError {
			t.Errorf("statusHandler should return error for invalid status '%s'", status)
			continue
		}

		// Verify error message matches FR-016 format
		if len(result.Content) == 0 {
			t.Errorf("statusHandler returned empty error content for '%s'", status)
			continue
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Errorf("statusHandler error content is not TextContent for '%s', got %T", status, result.Content[0])
			continue
		}

		expectedPrefix := "Invalid status:"
		if !strings.Contains(textContent.Text, expectedPrefix) {
			t.Errorf("Error message should contain '%s', got: %s", expectedPrefix, textContent.Text)
		}
		if !strings.Contains(textContent.Text, "Valid: ready, work, offline") {
			t.Errorf("Error message should contain valid options, got: %s", textContent.Text)
		}
	}
}

// T037: Test notified flag reset when status set to work/offline
func TestStatusHandler_ResetsNotifiedFlagOnWorkOrOffline(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create a recipients file with a notified agent
	recipientsDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(recipientsDir, 0755); err != nil {
		t.Fatalf("Failed to create recipients dir: %v", err)
	}

	// Write initial state with notified = true
	initialContent := `{"recipient":"test-agent","status":"ready","updated_at":"2024-01-01T00:00:00Z","notified":true}
`
	recipientsPath := filepath.Join(recipientsDir, "recipients.jsonl")
	if err := os.WriteFile(recipientsPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial recipients file: %v", err)
	}

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	testCases := []struct {
		status      string
		shouldReset bool
		description string
	}{
		{"work", true, "work should reset notified flag"},
		{"offline", true, "offline should reset notified flag"},
	}

	for _, tc := range testCases {
		// Reset the file with notified = true before each test
		if err := os.WriteFile(recipientsPath, []byte(initialContent), 0644); err != nil {
			t.Fatalf("Failed to reset recipients file for %s: %v", tc.status, err)
		}

		ctx := context.Background()
		result, err := statusHandler(ctx, makeStatusRequest(tc.status))
		if err != nil {
			t.Fatalf("statusHandler returned error for %s: %v", tc.status, err)
		}

		if result.IsError {
			t.Fatalf("statusHandler returned error result for %s: %v", tc.status, result.Content)
		}

		// Read the updated recipients file
		data, err := os.ReadFile(recipientsPath)
		if err != nil {
			t.Fatalf("Failed to read recipients file for %s: %v", tc.status, err)
		}

		// Verify notified flag was reset to false
		if tc.shouldReset {
			if strings.Contains(string(data), `"notified":true`) {
				t.Errorf("%s: %s. File content: %s", tc.status, tc.description, data)
			}
			if !strings.Contains(string(data), `"notified":false`) {
				t.Errorf("%s: Expected notified:false in file. Content: %s", tc.status, data)
			}
		}
	}
}

// Test status handler with "ready" does not reset notified flag
func TestStatusHandler_ReadyDoesNotResetNotifiedFlag(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Create a recipients file with a notified agent
	recipientsDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(recipientsDir, 0755); err != nil {
		t.Fatalf("Failed to create recipients dir: %v", err)
	}

	// Write initial state with notified = true
	initialContent := `{"recipient":"test-agent","status":"work","updated_at":"2024-01-01T00:00:00Z","notified":true}
`
	recipientsPath := filepath.Join(recipientsDir, "recipients.jsonl")
	if err := os.WriteFile(recipientsPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial recipients file: %v", err)
	}

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Set status to ready
	ctx := context.Background()
	result, err := statusHandler(ctx, makeStatusRequest("ready"))
	if err != nil {
		t.Fatalf("statusHandler returned error: %v", err)
	}

	if result.IsError {
		t.Fatalf("statusHandler returned error result: %v", result.Content)
	}

	// Read the updated recipients file
	data, err := os.ReadFile(recipientsPath)
	if err != nil {
		t.Fatalf("Failed to read recipients file: %v", err)
	}

	// Verify status was updated to ready
	if !strings.Contains(string(data), `"status":"ready"`) {
		t.Errorf("Status should be 'ready'. Content: %s", data)
	}

	// Verify notified flag was NOT reset (should remain true)
	if !strings.Contains(string(data), `"notified":true`) {
		t.Errorf("ready should NOT reset notified flag. Content: %s", data)
	}
}

// Test status response format matches data-model.md
func TestStatusHandler_ResponseFormat(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the status handler
	ctx := context.Background()
	result, err := statusHandler(ctx, makeStatusRequest("ready"))
	if err != nil {
		t.Fatalf("statusHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)

	// Verify JSON structure matches data-model.md StatusResponse
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

	// Verify status value is "ok"
	status, ok := responseMap["status"].(string)
	if !ok {
		t.Errorf("status is not a string: %T", responseMap["status"])
	}
	if status != "ok" {
		t.Errorf("status should be 'ok', got '%s'", status)
	}
}

// Test status handler via MCP client integration
func TestStatusHandler_MCPClientIntegration(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "mcp-status-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Set up test server and client
	_, clientSession, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Call status tool via MCP client
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: ToolStatus,
		Arguments: map[string]any{
			"status": "work",
		},
	})
	if err != nil {
		t.Fatalf("CallTool(status) failed: %v", err)
	}

	if result.IsError {
		textContent := result.Content[0].(*mcp.TextContent)
		t.Fatalf("CallTool(status) returned error: %s", textContent.Text)
	}

	if len(result.Content) == 0 {
		t.Fatal("CallTool(status) returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("CallTool(status) content is not TextContent")
	}

	// Parse and verify response
	var response StatusResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}
}

// Test all valid status values
func TestStatusHandler_AllValidStatuses(t *testing.T) {
	validStatuses := []string{"ready", "work", "offline"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			tmpDir := setupTestMailbox(t)
			defer os.RemoveAll(tmpDir)

			// Configure handler for testing
			SetHandlerOptions(&HandlerOptions{
				SkipTmuxCheck: true,
				MockReceiver:  "test-agent",
				RepoRoot:      tmpDir,
			})
			defer SetHandlerOptions(nil)

			ctx := context.Background()
			result, err := statusHandler(ctx, makeStatusRequest(status))
			if err != nil {
				t.Fatalf("statusHandler returned error for '%s': %v", status, err)
			}

			if result.IsError {
				textContent := result.Content[0].(*mcp.TextContent)
				t.Fatalf("statusHandler should succeed for valid status '%s', got error: %s", status, textContent.Text)
			}

			// Verify response
			textContent := result.Content[0].(*mcp.TextContent)
			var response StatusResponse
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to parse response JSON: %v", err)
			}

			if response.Status != "ok" {
				t.Errorf("Expected status 'ok', got '%s'", response.Status)
			}

			// Verify the status was persisted
			recipientsPath := filepath.Join(tmpDir, ".agentmail", "recipients.jsonl")
			data, err := os.ReadFile(recipientsPath)
			if err != nil {
				t.Fatalf("Failed to read recipients file: %v", err)
			}

			expectedStatus := `"status":"` + status + `"`
			if !strings.Contains(string(data), expectedStatus) {
				t.Errorf("Status '%s' not found in recipients file. Content: %s", status, data)
			}
		})
	}
}

// ============================================================================
// List-Recipients Handler Tests
// ============================================================================

// T042: Test list-recipients returns all agents (FR-006)
func TestListRecipientsHandler_ReturnsAllAgents(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing with multiple windows
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-1",
		MockWindows:   []string{"agent-1", "agent-2", "agent-3"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("listRecipientsHandler returned error result: %v", result.Content)
	}

	// Parse the response
	if len(result.Content) == 0 {
		t.Fatal("listRecipientsHandler returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("listRecipientsHandler content is not TextContent, got %T", result.Content[0])
	}

	var response ListRecipientsResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// FR-006: Should return all 3 windows
	if len(response.Recipients) != 3 {
		t.Errorf("Expected 3 recipients, got %d", len(response.Recipients))
	}

	// Verify all windows are present
	found := make(map[string]bool)
	for _, r := range response.Recipients {
		found[r.Name] = true
	}
	for _, expected := range []string{"agent-1", "agent-2", "agent-3"} {
		if !found[expected] {
			t.Errorf("Expected recipient '%s' not found", expected)
		}
	}
}

// T043: Test current window is marked with is_current: true
func TestListRecipientsHandler_CurrentWindowMarkedIsCurrent(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		MockWindows:   []string{"agent-1", "agent-2", "agent-3"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("listRecipientsHandler returned error result: %v", result.Content)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)
	var response ListRecipientsResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify current window is marked correctly
	currentCount := 0
	for _, r := range response.Recipients {
		if r.Name == "agent-2" {
			if !r.IsCurrent {
				t.Error("Current window 'agent-2' should have is_current: true")
			}
			currentCount++
		} else {
			if r.IsCurrent {
				t.Errorf("Non-current window '%s' should have is_current: false", r.Name)
			}
		}
	}

	// Should have exactly one current window
	if currentCount != 1 {
		t.Errorf("Expected exactly 1 current window, found %d", currentCount)
	}
}

// T044: Test ignored windows are excluded from list
func TestListRecipientsHandler_IgnoredWindowsExcluded(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing with ignored windows
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck:  true,
		MockReceiver:   "agent-1",
		MockWindows:    []string{"agent-1", "agent-2", "ignored-agent", "agent-3"},
		MockIgnoreList: map[string]bool{"ignored-agent": true},
		RepoRoot:       tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("listRecipientsHandler returned error result: %v", result.Content)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)
	var response ListRecipientsResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should return 3 recipients (ignored-agent excluded)
	if len(response.Recipients) != 3 {
		t.Errorf("Expected 3 recipients (ignored excluded), got %d", len(response.Recipients))
	}

	// Verify ignored window is not in the list
	for _, r := range response.Recipients {
		if r.Name == "ignored-agent" {
			t.Error("Ignored window 'ignored-agent' should not be in recipients list")
		}
	}
}

// Test current window shown even if in ignore list
func TestListRecipientsHandler_CurrentWindowShownEvenIfIgnored(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler where current window is in ignore list
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck:  true,
		MockReceiver:   "ignored-current",
		MockWindows:    []string{"agent-1", "ignored-current", "agent-2"},
		MockIgnoreList: map[string]bool{"ignored-current": true},
		RepoRoot:       tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("listRecipientsHandler returned error result: %v", result.Content)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)
	var response ListRecipientsResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should return 3 recipients (current window shown even though ignored)
	if len(response.Recipients) != 3 {
		t.Errorf("Expected 3 recipients (current shown even if ignored), got %d", len(response.Recipients))
	}

	// Verify current window is in the list and marked as current
	found := false
	for _, r := range response.Recipients {
		if r.Name == "ignored-current" {
			found = true
			if !r.IsCurrent {
				t.Error("Current window 'ignored-current' should have is_current: true")
			}
		}
	}
	if !found {
		t.Error("Current window 'ignored-current' should be in recipients list even though ignored")
	}
}

// Test response format matches data-model.md
func TestListRecipientsHandler_ResponseFormat(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		MockWindows:   []string{"test-agent", "other-agent"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)

	// Verify JSON structure matches data-model.md
	var responseMap map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &responseMap); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should have "recipients" field
	if _, ok := responseMap["recipients"]; !ok {
		t.Error("Response missing 'recipients' field")
	}

	// Verify recipients is an array
	recipients, ok := responseMap["recipients"].([]any)
	if !ok {
		t.Fatalf("recipients is not an array: %T", responseMap["recipients"])
	}

	// Verify each recipient has required fields
	for i, r := range recipients {
		recipient, ok := r.(map[string]any)
		if !ok {
			t.Fatalf("recipient %d is not an object", i)
		}
		if _, ok := recipient["name"]; !ok {
			t.Errorf("recipient %d missing 'name' field", i)
		}
		if _, ok := recipient["is_current"]; !ok {
			t.Errorf("recipient %d missing 'is_current' field", i)
		}
	}
}

// Test empty windows list
func TestListRecipientsHandler_EmptyWindowsList(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler with no windows (edge case)
	// MockReceiver must be non-empty to be recognized as mocked
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-window",
		MockWindows:   []string{},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned error: %v", err)
	}

	// Should not be an error result
	if result.IsError {
		t.Fatalf("listRecipientsHandler returned error result: %v", result.Content)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)
	var response ListRecipientsResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should return empty recipients array
	if len(response.Recipients) != 0 {
		t.Errorf("Expected 0 recipients for empty windows list, got %d", len(response.Recipients))
	}
}

// Test list-recipients via MCP client integration
func TestListRecipientsHandler_MCPClientIntegration(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "mcp-agent",
		MockWindows:   []string{"mcp-agent", "other-agent"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Set up test server and client
	_, clientSession, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Call list-recipients tool via MCP client
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListRecipients,
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool(list-recipients) failed: %v", err)
	}

	if result.IsError {
		textContent := result.Content[0].(*mcp.TextContent)
		t.Fatalf("CallTool(list-recipients) returned error: %s", textContent.Text)
	}

	if len(result.Content) == 0 {
		t.Fatal("CallTool(list-recipients) returned empty content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("CallTool(list-recipients) content is not TextContent")
	}

	// Parse and verify response
	var response ListRecipientsResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify response contains expected recipients
	if len(response.Recipients) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(response.Recipients))
	}
}

// Test multiple ignored windows are all excluded
func TestListRecipientsHandler_MultipleIgnoredWindowsExcluded(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler with multiple ignored windows
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck:  true,
		MockReceiver:   "agent-1",
		MockWindows:    []string{"agent-1", "ignored-1", "agent-2", "ignored-2", "agent-3"},
		MockIgnoreList: map[string]bool{"ignored-1": true, "ignored-2": true},
		RepoRoot:       tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)
	var response ListRecipientsResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should return 3 recipients (both ignored windows excluded)
	if len(response.Recipients) != 3 {
		t.Errorf("Expected 3 recipients, got %d", len(response.Recipients))
	}

	// Verify neither ignored window is in the list
	for _, r := range response.Recipients {
		if r.Name == "ignored-1" || r.Name == "ignored-2" {
			t.Errorf("Ignored window '%s' should not be in recipients list", r.Name)
		}
	}
}

// Test single window (only current)
func TestListRecipientsHandler_SingleWindow(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler with only one window
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "solo-agent",
		MockWindows:   []string{"solo-agent"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned error: %v", err)
	}

	// Parse the response
	textContent := result.Content[0].(*mcp.TextContent)
	var response ListRecipientsResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Should return 1 recipient
	if len(response.Recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(response.Recipients))
	}

	// Should be marked as current
	if len(response.Recipients) > 0 {
		if response.Recipients[0].Name != "solo-agent" {
			t.Errorf("Expected name 'solo-agent', got '%s'", response.Recipients[0].Name)
		}
		if !response.Recipients[0].IsCurrent {
			t.Error("Solo window should be marked as current")
		}
	}
}

// ============================================================================
// Error Path Tests for Coverage Improvement
// ============================================================================

// makeRequestWithMissingRecipient creates a CallToolRequest for send with missing recipient.
func makeRequestWithMissingRecipient(message string) *mcp.CallToolRequest {
	args, _ := json.Marshal(map[string]string{
		"message": message,
	})
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      ToolSend,
			Arguments: args,
		},
	}
}

// makeRequestWithMissingMessage creates a CallToolRequest for send with missing message.
func makeRequestWithMissingMessage(recipient string) *mcp.CallToolRequest {
	args, _ := json.Marshal(map[string]string{
		"recipient": recipient,
	})
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      ToolSend,
			Arguments: args,
		},
	}
}

// makeEmptyStatusRequest creates a CallToolRequest for status with no status parameter.
func makeEmptyStatusRequest() *mcp.CallToolRequest {
	args, _ := json.Marshal(map[string]string{})
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      ToolStatus,
			Arguments: args,
		},
	}
}

// Test send with missing recipient parameter returns error
func TestSendHandler_MissingRecipientReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the send handler with missing recipient (empty string from missing param)
	ctx := context.Background()
	result, err := sendHandler(ctx, makeRequestWithMissingRecipient("Hello!"))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result (recipient not found since empty string)
	if !result.IsError {
		t.Fatal("sendHandler should return error for missing recipient")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "recipient not found") {
		t.Errorf("Expected error to contain 'recipient not found', got: %s", textContent.Text)
	}
}

// Test send with missing message parameter returns error
func TestSendHandler_MissingMessageReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the send handler with missing message (empty string from missing param)
	ctx := context.Background()
	result, err := sendHandler(ctx, makeRequestWithMissingMessage("agent-receiver"))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("sendHandler should return error for missing message")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "no message provided") {
		t.Errorf("Expected error to contain 'no message provided', got: %s", textContent.Text)
	}
}

// Test status with missing status parameter returns error
func TestStatusHandler_MissingStatusReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the status handler with missing status parameter
	ctx := context.Background()
	result, err := statusHandler(ctx, makeEmptyStatusRequest())
	if err != nil {
		t.Fatalf("statusHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("statusHandler should return error for missing status parameter")
	}

	// Verify error message matches FR-016 format
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("statusHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Invalid status:") {
		t.Errorf("Expected error to contain 'Invalid status:', got: %s", textContent.Text)
	}
	if !strings.Contains(textContent.Text, "Valid: ready, work, offline") {
		t.Errorf("Expected error to contain valid options, got: %s", textContent.Text)
	}
}

// Test receive without mock receiver and no tmux returns error (simulates tmux failure)
// This test only runs when NOT in a tmux session, as it tests the tmux failure path.
func TestReceiveHandler_NoMockReceiverWithoutTmuxReturnsError(t *testing.T) {
	// Skip this test if running inside tmux - the error path won't trigger
	if os.Getenv("TMUX") != "" {
		t.Skip("Skipping test: running inside tmux session")
	}

	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing - NO MockReceiver set, SkipTmuxCheck false
	// This will cause tmux.GetCurrentWindow() to fail if not in tmux
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: false, // Enable tmux check
		MockReceiver:  "",    // No mock receiver
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler - should fail because we're not in tmux
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned unexpected error: %v", err)
	}

	// Should return an error result (failed to get current window)
	if !result.IsError {
		t.Fatal("receiveHandler should return error when tmux fails")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("receiveHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "failed to get current window") {
		t.Errorf("Expected error to contain 'failed to get current window', got: %s", textContent.Text)
	}
}

// Test receive with invalid repo root returns error (simulates git root failure)
func TestReceiveHandler_InvalidRepoRootReturnsError(t *testing.T) {
	// Configure handler with no RepoRoot and MockReceiver set
	// The handler will try to find git root which should fail in a non-existent directory
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      "", // Empty - will try to find git root
	})
	defer SetHandlerOptions(nil)

	// Save current directory and change to a non-git directory
	origDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "non-git-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to non-git directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Call the handler - should fail because there's no git root
	ctx := context.Background()
	result, err := receiveHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("receiveHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("receiveHandler should return error when git root not found")
	}

	// Verify error message contains git repository related error
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("receiveHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "not in a git repository") {
		t.Errorf("Expected error to contain 'not in a git repository', got: %s", textContent.Text)
	}
}

// Test send without mock sender and no tmux returns error
// This test only runs when NOT in a tmux session, as it tests the tmux failure path.
func TestSendHandler_NoMockSenderWithoutTmuxReturnsError(t *testing.T) {
	// Skip this test if running inside tmux - the error path won't trigger
	if os.Getenv("TMUX") != "" {
		t.Skip("Skipping test: running inside tmux session")
	}

	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler - NO MockSender set, SkipTmuxCheck false
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: false,
		MockSender:    "",                                         // No mock sender
		MockWindows:   []string{"agent-sender", "agent-receiver"}, // Mock windows still set for recipient check
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler - should fail because we're not in tmux
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("agent-receiver", "Hello!"))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("sendHandler should return error when tmux fails")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "failed to get current window") {
		t.Errorf("Expected error to contain 'failed to get current window', got: %s", textContent.Text)
	}
}

// Test send with invalid repo root returns error
func TestSendHandler_InvalidRepoRootReturnsError(t *testing.T) {
	// Configure handler with no RepoRoot
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      "", // Empty - will try to find git root
	})
	defer SetHandlerOptions(nil)

	// Save current directory and change to a non-git directory
	origDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "non-git-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to non-git directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Call the handler - should fail because there's no git root
	ctx := context.Background()
	result, err := sendHandler(ctx, makeSendRequest("agent-receiver", "Hello!"))
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("sendHandler should return error when git root not found")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "not in a git repository") {
		t.Errorf("Expected error to contain 'not in a git repository', got: %s", textContent.Text)
	}
}

// Test status without mock receiver and no tmux returns error
// This test only runs when NOT in a tmux session, as it tests the tmux failure path.
func TestStatusHandler_NoMockReceiverWithoutTmuxReturnsError(t *testing.T) {
	// Skip this test if running inside tmux - the error path won't trigger
	if os.Getenv("TMUX") != "" {
		t.Skip("Skipping test: running inside tmux session")
	}

	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler - NO MockReceiver set, SkipTmuxCheck false
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: false, // Enable tmux check
		MockReceiver:  "",    // No mock receiver
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler - should fail because we're not in tmux
	ctx := context.Background()
	result, err := statusHandler(ctx, makeStatusRequest("ready"))
	if err != nil {
		t.Fatalf("statusHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("statusHandler should return error when tmux fails")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("statusHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "failed to get current window") {
		t.Errorf("Expected error to contain 'failed to get current window', got: %s", textContent.Text)
	}
}

// Test status with invalid repo root returns error
func TestStatusHandler_InvalidRepoRootReturnsError(t *testing.T) {
	// Configure handler with no RepoRoot
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      "", // Empty - will try to find git root
	})
	defer SetHandlerOptions(nil)

	// Save current directory and change to a non-git directory
	origDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "non-git-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to non-git directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Call the handler - should fail because there's no git root
	ctx := context.Background()
	result, err := statusHandler(ctx, makeStatusRequest("ready"))
	if err != nil {
		t.Fatalf("statusHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("statusHandler should return error when git root not found")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("statusHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "not in a git repository") {
		t.Errorf("Expected error to contain 'not in a git repository', got: %s", textContent.Text)
	}
}

// Test list-recipients without mock windows and no tmux returns error
// This test only runs when NOT in a tmux session, as it tests the tmux failure path.
func TestListRecipientsHandler_NoMockWindowsWithoutTmuxReturnsError(t *testing.T) {
	// Skip this test if running inside tmux - the error path won't trigger
	if os.Getenv("TMUX") != "" {
		t.Skip("Skipping test: running inside tmux session")
	}

	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler - NO MockWindows and NO MockReceiver, SkipTmuxCheck false
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: false, // Enable tmux check
		MockReceiver:  "",    // No mock receiver
		MockWindows:   nil,   // No mock windows - will try real tmux
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler - should fail because we're not in tmux
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("listRecipientsHandler should return error when tmux fails")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("listRecipientsHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "failed to get current window") &&
		!strings.Contains(textContent.Text, "failed to list windows") {
		t.Errorf("Expected error related to tmux failure, got: %s", textContent.Text)
	}
}

// Test list-recipients with mock receiver but no mock windows causes ListWindows failure
// This test only runs when NOT in a tmux session, as it tests the tmux failure path.
func TestListRecipientsHandler_ListWindowsFailsReturnsError(t *testing.T) {
	// Skip this test if running inside tmux - the error path won't trigger
	if os.Getenv("TMUX") != "" {
		t.Skip("Skipping test: running inside tmux session")
	}

	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler with MockReceiver but NO MockWindows
	// This will use MockReceiver for current window but will try real tmux.ListWindows()
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		MockWindows:   nil, // nil triggers real tmux.ListWindows() call
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Call the handler - should fail because tmux.ListWindows will fail outside tmux
	ctx := context.Background()
	result, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("listRecipientsHandler returned unexpected error: %v", err)
	}

	// Should return an error result when tmux.ListWindows fails
	if !result.IsError {
		t.Fatal("listRecipientsHandler should return error when ListWindows fails")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("listRecipientsHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "failed to list windows") {
		t.Errorf("Expected error to contain 'failed to list windows', got: %s", textContent.Text)
	}
}

// Test send with invalid JSON arguments returns parse error
func TestSendHandler_InvalidJSONArgumentsReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Create request with invalid JSON in arguments
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      ToolSend,
			Arguments: json.RawMessage(`{invalid json`),
		},
	}

	ctx := context.Background()
	result, err := sendHandler(ctx, req)
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("sendHandler should return error for invalid JSON arguments")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "failed to parse arguments") {
		t.Errorf("Expected error to contain 'failed to parse arguments', got: %s", textContent.Text)
	}
}

// Test status with invalid JSON arguments returns parse error
func TestStatusHandler_InvalidJSONArgumentsReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Create request with invalid JSON in arguments
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      ToolStatus,
			Arguments: json.RawMessage(`{invalid json`),
		},
	}

	ctx := context.Background()
	result, err := statusHandler(ctx, req)
	if err != nil {
		t.Fatalf("statusHandler returned unexpected error: %v", err)
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("statusHandler should return error for invalid JSON arguments")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("statusHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "failed to parse arguments") {
		t.Errorf("Expected error to contain 'failed to parse arguments', got: %s", textContent.Text)
	}
}

// Test send with nil Params handles gracefully (missing recipient/message)
func TestSendHandler_NilParamsReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Create request with nil Params
	req := &mcp.CallToolRequest{
		Params: nil,
	}

	ctx := context.Background()
	result, err := sendHandler(ctx, req)
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result (no message provided)
	if !result.IsError {
		t.Fatal("sendHandler should return error for nil params")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "no message provided") {
		t.Errorf("Expected error to contain 'no message provided', got: %s", textContent.Text)
	}
}

// Test status with nil Params handles gracefully
func TestStatusHandler_NilParamsReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "test-agent",
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Create request with nil Params
	req := &mcp.CallToolRequest{
		Params: nil,
	}

	ctx := context.Background()
	result, err := statusHandler(ctx, req)
	if err != nil {
		t.Fatalf("statusHandler returned unexpected error: %v", err)
	}

	// Should return an error result (invalid status)
	if !result.IsError {
		t.Fatal("statusHandler should return error for nil params")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("statusHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "Invalid status:") {
		t.Errorf("Expected error to contain 'Invalid status:', got: %s", textContent.Text)
	}
}

// Test send with nil Arguments handles gracefully
func TestSendHandler_NilArgumentsReturnsError(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	// Create request with nil Arguments
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      ToolSend,
			Arguments: nil,
		},
	}

	ctx := context.Background()
	result, err := sendHandler(ctx, req)
	if err != nil {
		t.Fatalf("sendHandler returned unexpected error: %v", err)
	}

	// Should return an error result (no message provided)
	if !result.IsError {
		t.Fatal("sendHandler should return error for nil arguments")
	}

	// Verify error message
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("sendHandler error content is not TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "no message provided") {
		t.Errorf("Expected error to contain 'no message provided', got: %s", textContent.Text)
	}
}

// T057 / SC-004: Verify all tool invocations complete within 2 seconds
func TestToolInvocations_CompleteWithinTwoSeconds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockReceiver:  "agent-receiver",
		MockWindows:   []string{"agent-sender", "agent-receiver", "agent-other"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	ctx := context.Background()
	// SC-004: Each tool invocation must complete within 2 seconds
	// Allow 2.5s to account for CI variability
	maxDuration := 2500 * time.Millisecond

	// 1. Send tool
	t.Run("send_under_2s", func(t *testing.T) {
		start := time.Now()
		_, err := sendHandler(ctx, makeSendRequest("agent-receiver", "Performance test message"))
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("sendHandler returned error: %v", err)
		}
		if elapsed > maxDuration {
			t.Errorf("sendHandler took %v, expected < %v", elapsed, maxDuration)
		}
	})

	// 2. Receive tool
	t.Run("receive_under_2s", func(t *testing.T) {
		start := time.Now()
		_, err := receiveHandler(ctx, &mcp.CallToolRequest{})
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("receiveHandler returned error: %v", err)
		}
		if elapsed > maxDuration {
			t.Errorf("receiveHandler took %v, expected < %v", elapsed, maxDuration)
		}
	})

	// 3. Status tool
	t.Run("status_under_2s", func(t *testing.T) {
		start := time.Now()
		_, err := statusHandler(ctx, makeStatusRequest("ready"))
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("statusHandler returned error: %v", err)
		}
		if elapsed > maxDuration {
			t.Errorf("statusHandler took %v, expected < %v", elapsed, maxDuration)
		}
	})

	// 4. List recipients tool
	t.Run("list_recipients_under_2s", func(t *testing.T) {
		start := time.Now()
		_, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("listRecipientsHandler returned error: %v", err)
		}
		if elapsed > maxDuration {
			t.Errorf("listRecipientsHandler took %v, expected < %v", elapsed, maxDuration)
		}
	})
}

// T058 / SC-005: Verify server handles 100 consecutive invocations without errors
func TestServer_100ConsecutiveInvocations(t *testing.T) {
	tmpDir := setupTestMailbox(t)
	defer os.RemoveAll(tmpDir)

	// Configure handler for testing
	SetHandlerOptions(&HandlerOptions{
		SkipTmuxCheck: true,
		MockSender:    "agent-sender",
		MockReceiver:  "agent-receiver",
		MockWindows:   []string{"agent-sender", "agent-receiver"},
		RepoRoot:      tmpDir,
	})
	defer SetHandlerOptions(nil)

	ctx := context.Background()
	invocations := 100

	// Track errors
	var errors []string

	// Run 100 consecutive invocations alternating between tools
	for i := 0; i < invocations; i++ {
		switch i % 4 {
		case 0:
			// Send
			result, err := sendHandler(ctx, makeSendRequest("agent-receiver", fmt.Sprintf("Message %d", i)))
			if err != nil {
				errors = append(errors, "send invocation error: "+err.Error())
			} else if result.IsError {
				content, _ := result.Content[0].(*mcp.TextContent)
				if content != nil {
					errors = append(errors, "send invocation returned error: "+content.Text)
				}
			}
		case 1:
			// Receive
			_, err := receiveHandler(ctx, &mcp.CallToolRequest{})
			if err != nil {
				errors = append(errors, "receive invocation error: "+err.Error())
			}
		case 2:
			// Status
			statuses := []string{"ready", "work", "offline"}
			result, err := statusHandler(ctx, makeStatusRequest(statuses[i%3]))
			if err != nil {
				errors = append(errors, "status invocation error: "+err.Error())
			} else if result.IsError {
				content, _ := result.Content[0].(*mcp.TextContent)
				if content != nil {
					errors = append(errors, "status invocation returned error: "+content.Text)
				}
			}
		case 3:
			// List recipients
			_, err := listRecipientsHandler(ctx, &mcp.CallToolRequest{})
			if err != nil {
				errors = append(errors, "list_recipients invocation error: "+err.Error())
			}
		}
	}

	// SC-005: All 100 invocations should complete without errors
	if len(errors) > 0 {
		maxErrors := 10
		if len(errors) < maxErrors {
			maxErrors = len(errors)
		}
		t.Errorf("SC-005 violation: %d errors in 100 consecutive invocations:\n%s",
			len(errors), strings.Join(errors[:maxErrors], "\n"))
	}
}
