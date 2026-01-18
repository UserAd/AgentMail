package mail

import (
	"crypto/rand"
	"math/big"
	"time"
)

// Message represents a communication between agents.
// T008: Message struct with JSON tags
type Message struct {
	ID        string    `json:"id"`                   // Short unique identifier (8 chars, base62)
	From      string    `json:"from"`                 // Sender tmux window name
	To        string    `json:"to"`                   // Recipient tmux window name
	Message   string    `json:"message"`              // Body text
	ReadFlag  bool      `json:"read_flag"`            // Read status (default: false)
	CreatedAt time.Time `json:"created_at,omitempty"` // Timestamp for age-based cleanup
}

// base62 character set for ID generation
const base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateID creates a unique 8-character base62 identifier.
// T009: Implement GenerateID function (crypto/rand, base62)
func GenerateID() (string, error) {
	const idLength = 8
	result := make([]byte, idLength)
	charsetLen := big.NewInt(int64(len(base62Chars)))

	for i := 0; i < idLength; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = base62Chars[num.Int64()]
	}

	return string(result), nil
}
