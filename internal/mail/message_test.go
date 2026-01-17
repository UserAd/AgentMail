package mail

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// T006: Tests for Message struct JSON marshaling

func TestMessage_JSONMarshal(t *testing.T) {
	msg := Message{
		ID:       "xK7mN2pQ",
		From:     "agent-1",
		To:       "agent-2",
		Message:  "Hello from agent-1",
		ReadFlag: false,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal Message: %v", err)
	}

	// Verify JSON structure
	jsonStr := string(data)

	// Check all required fields are present
	if !strings.Contains(jsonStr, `"id":"xK7mN2pQ"`) {
		t.Errorf("JSON should contain id field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"from":"agent-1"`) {
		t.Errorf("JSON should contain from field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"to":"agent-2"`) {
		t.Errorf("JSON should contain to field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"message":"Hello from agent-1"`) {
		t.Errorf("JSON should contain message field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"read_flag":false`) {
		t.Errorf("JSON should contain read_flag field, got: %s", jsonStr)
	}
}

func TestMessage_JSONUnmarshal(t *testing.T) {
	jsonStr := `{"id":"xK7mN2pQ","from":"agent-1","to":"agent-2","message":"Hello from agent-1","read_flag":false}`

	var msg Message
	err := json.Unmarshal([]byte(jsonStr), &msg)
	if err != nil {
		t.Fatalf("Failed to unmarshal Message: %v", err)
	}

	if msg.ID != "xK7mN2pQ" {
		t.Errorf("Expected ID 'xK7mN2pQ', got '%s'", msg.ID)
	}
	if msg.From != "agent-1" {
		t.Errorf("Expected From 'agent-1', got '%s'", msg.From)
	}
	if msg.To != "agent-2" {
		t.Errorf("Expected To 'agent-2', got '%s'", msg.To)
	}
	if msg.Message != "Hello from agent-1" {
		t.Errorf("Expected Message 'Hello from agent-1', got '%s'", msg.Message)
	}
	if msg.ReadFlag != false {
		t.Errorf("Expected ReadFlag false, got %v", msg.ReadFlag)
	}
}

func TestMessage_JSONRoundTrip(t *testing.T) {
	original := Message{
		ID:       "zM9oP4rS",
		From:     "agent-3",
		To:       "agent-1",
		Message:  "Meeting at 3pm?",
		ReadFlag: true,
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var restored Message
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare
	if original != restored {
		t.Errorf("Round-trip failed: original %+v != restored %+v", original, restored)
	}
}

func TestMessage_JSONWithSpecialCharacters(t *testing.T) {
	msg := Message{
		ID:       "aB3cD4eF",
		From:     "agent-1",
		To:       "agent-2",
		Message:  `Hello "world"! Line1\nLine2`,
		ReadFlag: false,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message with special chars: %v", err)
	}

	var restored Message
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal message with special chars: %v", err)
	}

	if msg != restored {
		t.Errorf("Round-trip with special chars failed: original %+v != restored %+v", msg, restored)
	}
}

// T007: Tests for unique ID generation (8-char base62)

func TestGenerateID_Length(t *testing.T) {
	id, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID failed: %v", err)
	}

	if len(id) != 8 {
		t.Errorf("Expected ID length 8, got %d: '%s'", len(id), id)
	}
}

func TestGenerateID_Base62Characters(t *testing.T) {
	// Base62 = [a-zA-Z0-9]
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Generate multiple IDs to increase confidence
	for i := 0; i < 100; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed on iteration %d: %v", i, err)
		}

		for _, c := range id {
			if !strings.ContainsRune(validChars, c) {
				t.Errorf("ID contains invalid character '%c' in '%s'", c, id)
			}
		}
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	// Generate many IDs and check for collisions
	ids := make(map[string]bool)
	numIDs := 1000

	for i := 0; i < numIDs; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed: %v", err)
		}

		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

// Tests for CreatedAt timestamp serialization

func TestMessage_JSONMarshal_WithCreatedAt(t *testing.T) {
	timestamp := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	msg := Message{
		ID:        "xK7mN2pQ",
		From:      "agent-1",
		To:        "agent-2",
		Message:   "Hello from agent-1",
		ReadFlag:  false,
		CreatedAt: timestamp,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal Message with CreatedAt: %v", err)
	}

	jsonStr := string(data)

	// Verify created_at field is present
	if !strings.Contains(jsonStr, `"created_at":"2024-06-15T10:30:00Z"`) {
		t.Errorf("JSON should contain created_at field with correct timestamp, got: %s", jsonStr)
	}
}

func TestMessage_JSONUnmarshal_WithCreatedAt(t *testing.T) {
	jsonStr := `{"id":"xK7mN2pQ","from":"agent-1","to":"agent-2","message":"Hello from agent-1","read_flag":false,"created_at":"2024-06-15T10:30:00Z"}`

	var msg Message
	err := json.Unmarshal([]byte(jsonStr), &msg)
	if err != nil {
		t.Fatalf("Failed to unmarshal Message with CreatedAt: %v", err)
	}

	expectedTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	if !msg.CreatedAt.Equal(expectedTime) {
		t.Errorf("Expected CreatedAt '%v', got '%v'", expectedTime, msg.CreatedAt)
	}
}

func TestMessage_JSONMarshal_WithoutCreatedAt(t *testing.T) {
	msg := Message{
		ID:       "xK7mN2pQ",
		From:     "agent-1",
		To:       "agent-2",
		Message:  "Hello from agent-1",
		ReadFlag: false,
		// CreatedAt is zero value (not set)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal Message without CreatedAt: %v", err)
	}

	// Note: In Go, time.Time zero value is NOT omitted by omitempty because
	// time.Time is a struct, not a pointer. The omitempty tag only works for
	// truly "empty" values (nil pointers, empty strings, etc.).
	// This is acceptable because:
	// 1. New messages will always have CreatedAt set via Append()
	// 2. Legacy messages without created_at can still be unmarshaled correctly
	// 3. The zero time (0001-01-01T00:00:00Z) is distinguishable from real timestamps

	// Verify the JSON can be unmarshaled successfully (core requirement)
	var restored Message
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal Message with zero CreatedAt: %v", err)
	}

	// Verify CreatedAt is zero in the restored message
	if !restored.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be zero, got '%v'", restored.CreatedAt)
	}
}

func TestMessage_JSONRoundTrip_WithCreatedAt(t *testing.T) {
	timestamp := time.Date(2024, 6, 15, 14, 45, 30, 0, time.UTC)
	original := Message{
		ID:        "zM9oP4rS",
		From:      "agent-3",
		To:        "agent-1",
		Message:   "Meeting at 3pm?",
		ReadFlag:  true,
		CreatedAt: timestamp,
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var restored Message
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare all fields including CreatedAt
	if original.ID != restored.ID {
		t.Errorf("ID mismatch: original %s != restored %s", original.ID, restored.ID)
	}
	if original.From != restored.From {
		t.Errorf("From mismatch: original %s != restored %s", original.From, restored.From)
	}
	if original.To != restored.To {
		t.Errorf("To mismatch: original %s != restored %s", original.To, restored.To)
	}
	if original.Message != restored.Message {
		t.Errorf("Message mismatch: original %s != restored %s", original.Message, restored.Message)
	}
	if original.ReadFlag != restored.ReadFlag {
		t.Errorf("ReadFlag mismatch: original %v != restored %v", original.ReadFlag, restored.ReadFlag)
	}
	if !original.CreatedAt.Equal(restored.CreatedAt) {
		t.Errorf("CreatedAt mismatch: original %v != restored %v", original.CreatedAt, restored.CreatedAt)
	}
}

func TestMessage_JSONUnmarshal_BackwardCompatibility(t *testing.T) {
	// Test that messages without created_at field (legacy format) can still be parsed
	jsonStr := `{"id":"xK7mN2pQ","from":"agent-1","to":"agent-2","message":"Legacy message","read_flag":false}`

	var msg Message
	err := json.Unmarshal([]byte(jsonStr), &msg)
	if err != nil {
		t.Fatalf("Failed to unmarshal legacy Message without CreatedAt: %v", err)
	}

	// CreatedAt should be zero value for legacy messages
	if !msg.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be zero for legacy message, got '%v'", msg.CreatedAt)
	}

	// Other fields should still be correct
	if msg.ID != "xK7mN2pQ" {
		t.Errorf("Expected ID 'xK7mN2pQ', got '%s'", msg.ID)
	}
	if msg.From != "agent-1" {
		t.Errorf("Expected From 'agent-1', got '%s'", msg.From)
	}
	if msg.Message != "Legacy message" {
		t.Errorf("Expected Message 'Legacy message', got '%s'", msg.Message)
	}
}
