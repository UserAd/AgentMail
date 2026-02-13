package tmux

import (
	"testing"
)

// Tests for ParseAddress

func TestParseAddress_FullForm(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		currentSession string
		wantSession    string
		wantWindow     string
		wantPane       int
		wantErr        bool
	}{
		{
			name:           "basic full form",
			input:          "mysession:editor.0",
			currentSession: "",
			wantSession:    "mysession",
			wantWindow:     "editor",
			wantPane:       0,
			wantErr:        false,
		},
		{
			name:           "full form with pane 1",
			input:          "mysession:editor.1",
			currentSession: "",
			wantSession:    "mysession",
			wantWindow:     "editor",
			wantPane:       1,
			wantErr:        false,
		},
		{
			name:           "full form with dotted window name",
			input:          "mysession:my.app.0",
			currentSession: "",
			wantSession:    "mysession",
			wantWindow:     "my.app",
			wantPane:       0,
			wantErr:        false,
		},
		{
			name:           "full form multi-digit pane",
			input:          "s:w.12",
			currentSession: "",
			wantSession:    "s",
			wantWindow:     "w",
			wantPane:       12,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := ParseAddress(tt.input, tt.currentSession)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if addr.Session != tt.wantSession {
					t.Errorf("ParseAddress() Session = %v, want %v", addr.Session, tt.wantSession)
				}
				if addr.Window != tt.wantWindow {
					t.Errorf("ParseAddress() Window = %v, want %v", addr.Window, tt.wantWindow)
				}
				if addr.Pane != tt.wantPane {
					t.Errorf("ParseAddress() Pane = %v, want %v", addr.Pane, tt.wantPane)
				}
			}
		})
	}
}

func TestParseAddress_MediumForm(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		currentSession string
		wantSession    string
		wantWindow     string
		wantPane       int
		wantErr        bool
	}{
		{
			name:           "medium form",
			input:          ":editor.1",
			currentSession: "mysession",
			wantSession:    "mysession",
			wantWindow:     "editor",
			wantPane:       1,
			wantErr:        false,
		},
		{
			name:           "medium form with dotted window",
			input:          ":my.app.0",
			currentSession: "mysession",
			wantSession:    "mysession",
			wantWindow:     "my.app",
			wantPane:       0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := ParseAddress(tt.input, tt.currentSession)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if addr.Session != tt.wantSession {
					t.Errorf("ParseAddress() Session = %v, want %v", addr.Session, tt.wantSession)
				}
				if addr.Window != tt.wantWindow {
					t.Errorf("ParseAddress() Window = %v, want %v", addr.Window, tt.wantWindow)
				}
				if addr.Pane != tt.wantPane {
					t.Errorf("ParseAddress() Pane = %v, want %v", addr.Pane, tt.wantPane)
				}
			}
		})
	}
}

func TestParseAddress_ShortForm(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		currentSession string
		wantSession    string
		wantWindow     string
		wantPane       int
		wantErr        bool
	}{
		{
			name:           "short form simple window",
			input:          "editor",
			currentSession: "",
			wantSession:    "",
			wantWindow:     "editor",
			wantPane:       -1,
			wantErr:        false,
		},
		{
			name:           "short form with dots in window name",
			input:          "my.app",
			currentSession: "",
			wantSession:    "",
			wantWindow:     "my.app",
			wantPane:       -1,
			wantErr:        false,
		},
		{
			name:           "short form with dots like logs.1",
			input:          "logs.1",
			currentSession: "",
			wantSession:    "",
			wantWindow:     "logs.1",
			wantPane:       -1,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := ParseAddress(tt.input, tt.currentSession)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if addr.Session != tt.wantSession {
					t.Errorf("ParseAddress() Session = %v, want %v", addr.Session, tt.wantSession)
				}
				if addr.Window != tt.wantWindow {
					t.Errorf("ParseAddress() Window = %v, want %v", addr.Window, tt.wantWindow)
				}
				if addr.Pane != tt.wantPane {
					t.Errorf("ParseAddress() Pane = %v, want %v", addr.Pane, tt.wantPane)
				}
			}
		})
	}
}

func TestParseAddress_Invalid(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		currentSession string
	}{
		{
			name:           "empty string",
			input:          "",
			currentSession: "",
		},
		{
			name:           "colon alone",
			input:          ":",
			currentSession: "",
		},
		{
			name:           "starts with dot",
			input:          ".1",
			currentSession: "",
		},
		{
			name:           "invalid pane number",
			input:          "session:window.notanumber",
			currentSession: "",
		},
		{
			name:           "empty pane after dot",
			input:          "session:window.",
			currentSession: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAddress(tt.input, tt.currentSession)
			if err == nil {
				t.Errorf("ParseAddress() should return error for invalid input %q", tt.input)
			}
		})
	}
}

// Tests for FormatAddress

func TestFormatAddress(t *testing.T) {
	tests := []struct {
		name string
		addr *PaneAddress
		want string
	}{
		{
			name: "basic address",
			addr: &PaneAddress{Session: "s", Window: "w", Pane: 0},
			want: "s:w.0",
		},
		{
			name: "address with dotted window",
			addr: &PaneAddress{Session: "mysession", Window: "my.app", Pane: 1},
			want: "mysession:my.app.1",
		},
		{
			name: "address with multi-digit pane",
			addr: &PaneAddress{Session: "s", Window: "w", Pane: 12},
			want: "s:w.12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAddress(tt.addr)
			if got != tt.want {
				t.Errorf("FormatAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Tests for SanitizeForFilename

func TestSanitizeForFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic address",
			input: "mysession:editor.0",
			want:  "mysession%3Aeditor%2E0",
		},
		{
			name:  "address with percent character",
			input: "my%session:win.0",
			want:  "my%25session%3Awin%2E0",
		},
		{
			name:  "address with dotted window name",
			input: "s:my.app.0",
			want:  "s%3Amy.app%2E0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeForFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeForFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Tests for UnsanitizeFromFilename

func TestUnsanitizeFromFilename_Roundtrip(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{
			name:    "basic address",
			address: "mysession:editor.0",
		},
		{
			name:    "address with percent",
			address: "my%session:win.0",
		},
		{
			name:    "address with dotted window",
			address: "s:my.app.0",
		},
		{
			name:    "complex address",
			address: "my%session:my.app.window.12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := SanitizeForFilename(tt.address)
			unsanitized := UnsanitizeFromFilename(sanitized)
			if unsanitized != tt.address {
				t.Errorf("Roundtrip failed: original=%q, sanitized=%q, unsanitized=%q", tt.address, sanitized, unsanitized)
			}
		})
	}
}

// Test for collision resistance

func TestSanitizeForFilename_NoCollisions(t *testing.T) {
	addr1 := "s_a:w.0"
	addr2 := "s:a_w.0"

	sanitized1 := SanitizeForFilename(addr1)
	sanitized2 := SanitizeForFilename(addr2)

	if sanitized1 == sanitized2 {
		t.Errorf("Collision detected: %q and %q both sanitize to %q", addr1, addr2, sanitized1)
	}
}
