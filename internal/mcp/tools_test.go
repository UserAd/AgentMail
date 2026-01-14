package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// testImpl is a test implementation for the MCP server.
var testImpl = &mcp.Implementation{Name: "agentmail-test", Version: "test"}

// setupTestServer creates a connected server and client for testing tools.
func setupTestServer(t *testing.T) (*Server, *mcp.ClientSession) {
	t.Helper()

	// Create server with tmux check skipped
	server, err := NewServer(&ServerOptions{
		SkipTmuxCheck: true,
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// Create in-memory transports
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	ctx := context.Background()
	_, err = server.MCPServer().Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("Server connect failed: %v", err)
	}

	// Create and connect client
	client := mcp.NewClient(testImpl, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Client connect failed: %v", err)
	}

	return server, clientSession
}

func TestRegisterTools_FourToolsExposed(t *testing.T) {
	// T010: Test that all 4 tools are registered
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()
	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(result.Tools) != 4 {
		t.Errorf("expected 4 tools, got %d", len(result.Tools))
	}

	// Verify all expected tools are present
	expectedTools := map[string]bool{
		ToolSend:           false,
		ToolReceive:        false,
		ToolStatus:         false,
		ToolListRecipients: false,
	}

	for _, tool := range result.Tools {
		if _, exists := expectedTools[tool.Name]; exists {
			expectedTools[tool.Name] = true
		} else {
			t.Errorf("unexpected tool: %s", tool.Name)
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("expected tool not found: %s", name)
		}
	}
}

func TestRegisterTools_EachToolHasDescription(t *testing.T) {
	// T011: Test that each tool has a description (FR-011)
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()
	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
	}
}

func TestRegisterTools_EachToolHasInputSchema(t *testing.T) {
	// T011: Test that each tool has an input schema (FR-011)
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()
	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.InputSchema == nil {
			t.Errorf("tool %q has nil input schema", tool.Name)
		}
	}
}

func TestSendTool_SchemaValidation(t *testing.T) {
	// T013: Test send tool schema has recipient and message parameters
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()
	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	var sendTool *mcp.Tool
	for _, tool := range result.Tools {
		if tool.Name == ToolSend {
			sendTool = tool
			break
		}
	}

	if sendTool == nil {
		t.Fatal("send tool not found")
	}

	// Verify description
	expectedDesc := "Send a message to another agent in a tmux window"
	if sendTool.Description != expectedDesc {
		t.Errorf("send tool description mismatch: got %q, want %q", sendTool.Description, expectedDesc)
	}

	// Verify input schema is present (schema is returned as map[string]any from SDK)
	schema, ok := sendTool.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("send tool input schema is not a map, got %T", sendTool.InputSchema)
	}

	// Verify schema type is object
	if schemaType, ok := schema["type"].(string); !ok || schemaType != "object" {
		t.Errorf("send tool schema type is not 'object': %v", schema["type"])
	}

	// Verify properties exist
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("send tool schema properties is not a map: %T", schema["properties"])
	}

	// Verify recipient property exists
	if _, ok := props["recipient"]; !ok {
		t.Error("send tool schema missing 'recipient' property")
	}

	// Verify message property exists
	if _, ok := props["message"]; !ok {
		t.Error("send tool schema missing 'message' property")
	}

	// Verify required fields
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("send tool schema required is not an array: %T", schema["required"])
	}

	hasRecipientRequired := false
	hasMessageRequired := false
	for _, r := range required {
		if str, ok := r.(string); ok {
			if str == "recipient" {
				hasRecipientRequired = true
			}
			if str == "message" {
				hasMessageRequired = true
			}
		}
	}

	if !hasRecipientRequired {
		t.Error("send tool schema missing 'recipient' in required fields")
	}
	if !hasMessageRequired {
		t.Error("send tool schema missing 'message' in required fields")
	}
}

func TestReceiveTool_SchemaValidation(t *testing.T) {
	// T014: Test receive tool schema has no required parameters
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()
	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	var receiveTool *mcp.Tool
	for _, tool := range result.Tools {
		if tool.Name == ToolReceive {
			receiveTool = tool
			break
		}
	}

	if receiveTool == nil {
		t.Fatal("receive tool not found")
	}

	// Verify description
	expectedDesc := "Read the oldest unread message from your mailbox"
	if receiveTool.Description != expectedDesc {
		t.Errorf("receive tool description mismatch: got %q, want %q", receiveTool.Description, expectedDesc)
	}

	// Verify input schema is present
	schema, ok := receiveTool.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("receive tool input schema is not a map, got %T", receiveTool.InputSchema)
	}

	// Verify schema type is object
	if schemaType, ok := schema["type"].(string); !ok || schemaType != "object" {
		t.Errorf("receive tool schema type is not 'object': %v", schema["type"])
	}
}

func TestStatusTool_SchemaValidation(t *testing.T) {
	// T015: Test status tool schema has status enum parameter
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()
	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	var statusTool *mcp.Tool
	for _, tool := range result.Tools {
		if tool.Name == ToolStatus {
			statusTool = tool
			break
		}
	}

	if statusTool == nil {
		t.Fatal("status tool not found")
	}

	// Verify description
	expectedDesc := "Set your agent's availability status"
	if statusTool.Description != expectedDesc {
		t.Errorf("status tool description mismatch: got %q, want %q", statusTool.Description, expectedDesc)
	}

	// Verify input schema is present
	schema, ok := statusTool.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("status tool input schema is not a map, got %T", statusTool.InputSchema)
	}

	// Verify schema type is object
	if schemaType, ok := schema["type"].(string); !ok || schemaType != "object" {
		t.Errorf("status tool schema type is not 'object': %v", schema["type"])
	}

	// Verify properties exist
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("status tool schema properties is not a map: %T", schema["properties"])
	}

	// Verify status property exists with enum
	statusProp, ok := props["status"].(map[string]any)
	if !ok {
		t.Fatal("status tool schema missing 'status' property")
	}

	// Verify enum values
	enumVal, ok := statusProp["enum"].([]any)
	if !ok {
		t.Fatalf("status property enum is not an array: %T", statusProp["enum"])
	}

	expectedEnums := map[string]bool{"ready": false, "work": false, "offline": false}
	for _, e := range enumVal {
		if str, ok := e.(string); ok {
			if _, exists := expectedEnums[str]; exists {
				expectedEnums[str] = true
			}
		}
	}

	for enumVal, found := range expectedEnums {
		if !found {
			t.Errorf("status enum missing value: %s", enumVal)
		}
	}

	// Verify status is required
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("status tool schema required is not an array: %T", schema["required"])
	}

	hasStatusRequired := false
	for _, r := range required {
		if str, ok := r.(string); ok && str == "status" {
			hasStatusRequired = true
			break
		}
	}

	if !hasStatusRequired {
		t.Error("status tool schema missing 'status' in required fields")
	}
}

func TestListRecipientsTool_SchemaValidation(t *testing.T) {
	// T016: Test list-recipients tool schema has no required parameters
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()
	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	var listRecipientsTool *mcp.Tool
	for _, tool := range result.Tools {
		if tool.Name == ToolListRecipients {
			listRecipientsTool = tool
			break
		}
	}

	if listRecipientsTool == nil {
		t.Fatal("list-recipients tool not found")
	}

	// Verify description
	expectedDesc := "List all available agents that can receive messages"
	if listRecipientsTool.Description != expectedDesc {
		t.Errorf("list-recipients tool description mismatch: got %q, want %q", listRecipientsTool.Description, expectedDesc)
	}

	// Verify input schema is present
	schema, ok := listRecipientsTool.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("list-recipients tool input schema is not a map, got %T", listRecipientsTool.InputSchema)
	}

	// Verify schema type is object
	if schemaType, ok := schema["type"].(string); !ok || schemaType != "object" {
		t.Errorf("list-recipients tool schema type is not 'object': %v", schema["type"])
	}
}

func TestToolDiscovery_Performance(t *testing.T) {
	// T018: Verify tool discovery completes within 1 second (SC-001)
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()

	start := time.Now()
	_, err := clientSession.ListTools(ctx, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	// SC-001: Tool discovery must complete within 1 second
	if elapsed > time.Second {
		t.Errorf("tool discovery took %v, expected < 1 second", elapsed)
	}
}

func TestToolHandlers_ReturnNotImplemented(t *testing.T) {
	// Test that stub tool handlers return "not implemented"
	// Note: receive and send tools are now implemented - tested in handlers_test.go
	_, clientSession := setupTestServer(t)
	defer clientSession.Close()

	ctx := context.Background()

	testCases := []struct {
		name      string
		toolName  string
		arguments map[string]any
	}{
		// send tool is now implemented - see handlers_test.go
		// receive tool is now implemented - see handlers_test.go
		{
			name:     "status tool",
			toolName: ToolStatus,
			arguments: map[string]any{
				"status": "ready",
			},
		},
		{
			name:      "list-recipients tool",
			toolName:  ToolListRecipients,
			arguments: map[string]any{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
				Name:      tc.toolName,
				Arguments: tc.arguments,
			})
			if err != nil {
				t.Fatalf("CallTool(%s) failed: %v", tc.toolName, err)
			}

			if len(result.Content) == 0 {
				t.Fatalf("CallTool(%s) returned empty content", tc.toolName)
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatalf("CallTool(%s) content is not TextContent", tc.toolName)
			}

			if textContent.Text != "not implemented" {
				t.Errorf("CallTool(%s) returned %q, expected 'not implemented'", tc.toolName, textContent.Text)
			}
		})
	}
}

func TestToolConstants(t *testing.T) {
	// Verify tool name constants match expected values
	if ToolSend != "send" {
		t.Errorf("ToolSend constant mismatch: got %q, want 'send'", ToolSend)
	}
	if ToolReceive != "receive" {
		t.Errorf("ToolReceive constant mismatch: got %q, want 'receive'", ToolReceive)
	}
	if ToolStatus != "status" {
		t.Errorf("ToolStatus constant mismatch: got %q, want 'status'", ToolStatus)
	}
	if ToolListRecipients != "list-recipients" {
		t.Errorf("ToolListRecipients constant mismatch: got %q, want 'list-recipients'", ToolListRecipients)
	}
}

func TestMaxMessageSize(t *testing.T) {
	// Verify MaxMessageSize constant is 64KB
	if MaxMessageSize != 65536 {
		t.Errorf("MaxMessageSize constant mismatch: got %d, want 65536", MaxMessageSize)
	}
}

func TestSendToolSchema_MaxLength(t *testing.T) {
	// Verify the send tool schema includes maxLength constraint for message
	schema := sendToolSchema()

	var schemaMap map[string]any
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		t.Fatalf("Failed to unmarshal send tool schema: %v", err)
	}

	props, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		t.Fatal("send tool schema missing properties")
	}

	messageProp, ok := props["message"].(map[string]any)
	if !ok {
		t.Fatal("send tool schema missing message property")
	}

	maxLength, ok := messageProp["maxLength"].(float64)
	if !ok {
		t.Fatal("send tool schema missing maxLength for message property")
	}

	if int(maxLength) != MaxMessageSize {
		t.Errorf("send tool message maxLength mismatch: got %d, want %d", int(maxLength), MaxMessageSize)
	}
}
