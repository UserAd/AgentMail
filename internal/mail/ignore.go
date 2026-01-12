package mail

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// FindGitRoot walks up the directory tree looking for a .git directory.
// Returns the path to the git repository root or an error if not in a git repository.
func FindGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("not in a git repository")
		}
		dir = parent
	}
}

// LoadIgnoreList reads the .agentmailignore file from the git root directory.
// It returns a map of window names to ignore for easy lookup.
// Per FR-016: If the file doesn't exist or is unreadable, returns nil (no error).
func LoadIgnoreList(gitRoot string) (map[string]bool, error) {
	path := filepath.Join(gitRoot, ".agentmailignore") // #nosec G304 - filename is a constant
	file, err := os.Open(path)                         // #nosec G304 - path is constructed from constant
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return nil, nil // Per FR-016: treat as if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	ignored := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			ignored[line] = true
		}
	}
	return ignored, scanner.Err()
}
