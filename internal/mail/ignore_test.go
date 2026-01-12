package mail

import (
	"os"
	"path/filepath"
	"testing"
)

// Tests for FindGitRoot function

func TestFindGitRoot_ReturnsCorrectPath(t *testing.T) {
	// Create temp dir as a fake git repo
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks: %v", err)
	}

	// Create .git directory to simulate git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create a subdirectory to test walking up
	subDir := filepath.Join(tmpDir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirs: %v", err)
	}

	// Save current working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(origDir)

	// Change to subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Test FindGitRoot
	root, err := FindGitRoot()
	if err != nil {
		t.Fatalf("FindGitRoot failed: %v", err)
	}

	if root != tmpDir {
		t.Errorf("Expected git root %s, got %s", tmpDir, root)
	}
}

func TestFindGitRoot_ErrorWhenNotInGitRepo(t *testing.T) {
	// Create temp dir without .git
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(origDir)

	// Change to temp directory (no .git)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Test FindGitRoot - should fail
	_, err = FindGitRoot()
	if err == nil {
		t.Error("FindGitRoot should return error when not in a git repo")
	}

	expectedErr := "not in a git repository"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// Tests for LoadIgnoreList function

func TestLoadIgnoreList_LoadsEntriesCorrectly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmailignore with entries
	content := `agent-1
agent-2
some-window
`
	ignorePath := filepath.Join(tmpDir, ".agentmailignore")
	if err := os.WriteFile(ignorePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write ignore file: %v", err)
	}

	ignored, err := LoadIgnoreList(tmpDir)
	if err != nil {
		t.Fatalf("LoadIgnoreList failed: %v", err)
	}

	if ignored == nil {
		t.Fatal("Expected non-nil map")
	}

	if len(ignored) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(ignored))
	}

	// Check all entries exist
	expectedEntries := []string{"agent-1", "agent-2", "some-window"}
	for _, entry := range expectedEntries {
		if !ignored[entry] {
			t.Errorf("Expected entry '%s' to be in ignore list", entry)
		}
	}
}

func TestLoadIgnoreList_HandlesMissingFileGracefully(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// No .agentmailignore file exists

	ignored, err := LoadIgnoreList(tmpDir)
	if err != nil {
		t.Errorf("LoadIgnoreList should not return error for missing file: %v", err)
	}

	if ignored != nil {
		t.Errorf("Expected nil map for missing file, got %v", ignored)
	}
}

func TestLoadIgnoreList_IgnoresEmptyAndWhitespaceLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmailignore with empty lines and whitespace
	content := `agent-1



agent-2
   agent-3
`
	ignorePath := filepath.Join(tmpDir, ".agentmailignore")
	if err := os.WriteFile(ignorePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write ignore file: %v", err)
	}

	ignored, err := LoadIgnoreList(tmpDir)
	if err != nil {
		t.Fatalf("LoadIgnoreList failed: %v", err)
	}

	if ignored == nil {
		t.Fatal("Expected non-nil map")
	}

	// Should only have 3 non-empty entries (with whitespace trimmed)
	if len(ignored) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(ignored))
	}

	// Check trimming worked
	if !ignored["agent-1"] {
		t.Error("Expected 'agent-1' in ignore list")
	}
	if !ignored["agent-2"] {
		t.Error("Expected 'agent-2' in ignore list")
	}
	if !ignored["agent-3"] {
		t.Error("Expected 'agent-3' (trimmed) in ignore list")
	}

	// Verify whitespace-only lines are not included
	if ignored[""] {
		t.Error("Empty string should not be in ignore list")
	}
	if ignored["   "] {
		t.Error("Whitespace-only entry should not be in ignore list")
	}
}

func TestLoadIgnoreList_HandlesUnreadableFileGracefully(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmailignore with no read permissions
	ignorePath := filepath.Join(tmpDir, ".agentmailignore")
	if err := os.WriteFile(ignorePath, []byte("agent-1\n"), 0000); err != nil {
		t.Fatalf("Failed to write ignore file: %v", err)
	}
	defer os.Chmod(ignorePath, 0644) // Ensure cleanup

	ignored, err := LoadIgnoreList(tmpDir)
	if err != nil {
		t.Errorf("LoadIgnoreList should not return error for unreadable file: %v", err)
	}

	if ignored != nil {
		t.Errorf("Expected nil map for unreadable file, got %v", ignored)
	}
}

func TestLoadIgnoreList_LookupFunctionality(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmailignore
	content := `ignored-window
another-ignored
`
	ignorePath := filepath.Join(tmpDir, ".agentmailignore")
	if err := os.WriteFile(ignorePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write ignore file: %v", err)
	}

	ignored, err := LoadIgnoreList(tmpDir)
	if err != nil {
		t.Fatalf("LoadIgnoreList failed: %v", err)
	}

	// Test lookup functionality
	if !ignored["ignored-window"] {
		t.Error("Should find 'ignored-window' in ignore list")
	}
	if !ignored["another-ignored"] {
		t.Error("Should find 'another-ignored' in ignore list")
	}
	if ignored["not-ignored"] {
		t.Error("Should not find 'not-ignored' in ignore list")
	}
}
