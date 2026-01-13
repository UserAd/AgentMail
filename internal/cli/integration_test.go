package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// extractMessageID extracts the message ID from send output format "Message #ID sent"
func extractMessageID(output string) string {
	output = strings.TrimSpace(output)
	// Format: "Message #ID sent"
	if strings.HasPrefix(output, "Message #") && strings.HasSuffix(output, " sent") {
		// Extract ID between "Message #" and " sent"
		return output[9 : len(output)-5]
	}
	return output
}

// T039: Integration test: send -> receive round-trip

func TestIntegration_SendReceiveRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Agent-1 sends a message to Agent-2
	var sendStdout, sendStderr bytes.Buffer
	sendExit := Send([]string{"agent-2", "Hello from agent-1!"}, nil, &sendStdout, &sendStderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if sendExit != 0 {
		t.Fatalf("Send failed with exit code %d: %s", sendExit, sendStderr.String())
	}

	// Get the message ID from send output (format: "Message #ID sent")
	messageID := extractMessageID(sendStdout.String())
	if len(messageID) != 8 {
		t.Fatalf("Expected 8-char message ID, got: %s", messageID)
	}

	// Agent-2 receives the message
	var recvStdout, recvStderr bytes.Buffer
	recvExit := Receive(&recvStdout, &recvStderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})

	if recvExit != 0 {
		t.Fatalf("Receive failed with exit code %d: %s", recvExit, recvStderr.String())
	}

	output := recvStdout.String()

	// Verify message content
	if !strings.Contains(output, "From: agent-1") {
		t.Errorf("Expected 'From: agent-1' in output, got: %s", output)
	}
	if !strings.Contains(output, "ID: "+messageID) {
		t.Errorf("Expected 'ID: %s' in output, got: %s", messageID, output)
	}
	if !strings.Contains(output, "Hello from agent-1!") {
		t.Errorf("Expected message body in output, got: %s", output)
	}

	// Verify subsequent receive shows no messages
	var recvStdout2, recvStderr2 bytes.Buffer
	recvExit2 := Receive(&recvStdout2, &recvStderr2, ReceiveOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})

	if recvExit2 != 0 {
		t.Fatalf("Second receive failed with exit code %d", recvExit2)
	}

	if !strings.Contains(recvStdout2.String(), "No unread messages") {
		t.Errorf("Expected 'No unread messages' after receiving, got: %s", recvStdout2.String())
	}
}

// T040: Integration test: FIFO ordering (send 3, receive 3)

func TestIntegration_FIFOOrdering(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Send 3 messages
	messages := []string{"First message", "Second message", "Third message"}
	var messageIDs []string

	for _, msg := range messages {
		var stdout, stderr bytes.Buffer
		exitCode := Send([]string{"agent-2", msg}, nil, &stdout, &stderr, SendOptions{
			SkipTmuxCheck: true,
			MockWindows:   []string{"agent-1", "agent-2"},
			MockSender:    "agent-1",
			RepoRoot:      tmpDir,
		})

		if exitCode != 0 {
			t.Fatalf("Send failed: %s", stderr.String())
		}
		messageIDs = append(messageIDs, extractMessageID(stdout.String()))
	}

	// Receive 3 messages and verify FIFO order
	for i, expectedMsg := range messages {
		var stdout, stderr bytes.Buffer
		exitCode := Receive(&stdout, &stderr, ReceiveOptions{
			SkipTmuxCheck: true,
			MockWindows:   []string{"agent-1", "agent-2"},
			MockReceiver:  "agent-2",
			RepoRoot:      tmpDir,
		})

		if exitCode != 0 {
			t.Fatalf("Receive %d failed: %s", i+1, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, expectedMsg) {
			t.Errorf("Receive %d: Expected message '%s', got: %s", i+1, expectedMsg, output)
		}
		if !strings.Contains(output, "ID: "+messageIDs[i]) {
			t.Errorf("Receive %d: Expected ID '%s' in output", i+1, messageIDs[i])
		}
	}

	// Verify no more messages
	var stdout, stderr bytes.Buffer
	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Fatalf("Final receive failed: %s", stderr.String())
	}

	if !strings.Contains(stdout.String(), "No unread messages") {
		t.Errorf("Expected 'No unread messages' after all received")
	}
}

// T041: Integration test: multi-agent file isolation (separate .jsonl per recipient)

func TestIntegration_MultiAgentFileIsolation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	windows := []string{"agent-1", "agent-2", "agent-3"}

	// Agent-1 sends to Agent-2
	var stdout1, stderr1 bytes.Buffer
	Send([]string{"agent-2", "Message for agent-2"}, nil, &stdout1, &stderr1, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   windows,
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
	})

	// Agent-1 sends to Agent-3
	var stdout2, stderr2 bytes.Buffer
	Send([]string{"agent-3", "Message for agent-3"}, nil, &stdout2, &stderr2, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   windows,
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
	})

	// Verify separate files exist
	agent2File := filepath.Join(mailDir, "agent-2.jsonl")
	agent3File := filepath.Join(mailDir, "agent-3.jsonl")

	if _, err := os.Stat(agent2File); os.IsNotExist(err) {
		t.Error("agent-2.jsonl should exist")
	}
	if _, err := os.Stat(agent3File); os.IsNotExist(err) {
		t.Error("agent-3.jsonl should exist")
	}

	// Verify agent-1 has no mail file (didn't receive any messages)
	agent1File := filepath.Join(mailDir, "agent-1.jsonl")
	if _, err := os.Stat(agent1File); !os.IsNotExist(err) {
		t.Error("agent-1.jsonl should NOT exist (no messages sent to agent-1)")
	}

	// Agent-2 receives their message
	var recv2Stdout, recv2Stderr bytes.Buffer
	Receive(&recv2Stdout, &recv2Stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockWindows:   windows,
		MockReceiver:  "agent-2",
		RepoRoot:      tmpDir,
	})

	if !strings.Contains(recv2Stdout.String(), "Message for agent-2") {
		t.Errorf("Agent-2 should receive their message, got: %s", recv2Stdout.String())
	}

	// Agent-3 receives their message
	var recv3Stdout, recv3Stderr bytes.Buffer
	Receive(&recv3Stdout, &recv3Stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockWindows:   windows,
		MockReceiver:  "agent-3",
		RepoRoot:      tmpDir,
	})

	if !strings.Contains(recv3Stdout.String(), "Message for agent-3") {
		t.Errorf("Agent-3 should receive their message, got: %s", recv3Stdout.String())
	}

	// Agent-1 should have no messages
	var recv1Stdout, recv1Stderr bytes.Buffer
	Receive(&recv1Stdout, &recv1Stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockWindows:   windows,
		MockReceiver:  "agent-1",
		RepoRoot:      tmpDir,
	})

	if !strings.Contains(recv1Stdout.String(), "No unread messages") {
		t.Errorf("Agent-1 should have no messages, got: %s", recv1Stdout.String())
	}
}
