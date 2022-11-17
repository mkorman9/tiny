package tiny

import (
	"crypto/rand"
	"encoding/hex"
)

// GetSecureRandomString generates a cryptographically secure string of given length.
// Result string is encoded to hex.
func GetSecureRandomString(lengthBytes uint) (string, error) {
	var bytes = make([]byte, lengthBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
