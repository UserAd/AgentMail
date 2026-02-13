package tmux

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PaneAddress represents a parsed tmux pane address.
type PaneAddress struct {
	Session string
	Window  string
	Pane    int
	Full    string
}

var (
	fullFormPattern   = regexp.MustCompile(`^([^:]+):(.+)\.(\d+)$`)
	mediumFormPattern = regexp.MustCompile(`^:(.+)\.(\d+)$`)
)

// ParseAddress parses a tmux pane address string.
func ParseAddress(input string, currentSession string) (*PaneAddress, error) {
	if input == "" {
		return nil, errors.New("empty address")
	}

	// Try full form: session:window.pane
	if matches := fullFormPattern.FindStringSubmatch(input); matches != nil {
		pane, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil, errors.New("invalid pane number")
		}
		addr := &PaneAddress{
			Session: matches[1],
			Window:  matches[2],
			Pane:    pane,
		}
		addr.Full = FormatAddress(addr)
		return addr, nil
	}

	// Try medium form: :window.pane
	if matches := mediumFormPattern.FindStringSubmatch(input); matches != nil {
		pane, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, errors.New("invalid pane number")
		}
		addr := &PaneAddress{
			Session: currentSession,
			Window:  matches[1],
			Pane:    pane,
		}
		addr.Full = FormatAddress(addr)
		return addr, nil
	}

	// Short form: window (no colon, no pane)
	if !strings.Contains(input, ":") && !strings.HasPrefix(input, ".") {
		return &PaneAddress{
			Session: "",
			Window:  input,
			Pane:    -1,
			Full:    input,
		}, nil
	}

	return nil, errors.New("invalid address format")
}

// FormatAddress formats a PaneAddress as a string.
func FormatAddress(addr *PaneAddress) string {
	return fmt.Sprintf("%s:%s.%d", addr.Session, addr.Window, addr.Pane)
}

// SanitizeForFilename converts a pane address to a filesystem-safe filename.
func SanitizeForFilename(address string) string {
	result := strings.ReplaceAll(address, "%", "%25")
	result = strings.ReplaceAll(result, ":", "%3A")
	lastDot := strings.LastIndex(result, ".")
	if lastDot != -1 {
		result = result[:lastDot] + "%2E" + result[lastDot+1:]
	}
	return result
}

// UnsanitizeFromFilename converts a sanitized filename back to a pane address.
func UnsanitizeFromFilename(filename string) string {
	result := strings.ReplaceAll(filename, "%2E", ".")
	result = strings.ReplaceAll(result, "%3A", ":")
	result = strings.ReplaceAll(result, "%25", "%")
	return result
}
