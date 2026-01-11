package mail

import (
	"encoding/json"
	"strings"
	"testing"
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
