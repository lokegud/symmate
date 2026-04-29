package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 3
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	keyLen        = 32 // 256-bit key for XChaCha20-Poly1305
	saltLen       = 16
)

// DeriveKey derives a 32-byte key from passphrase using Argon2id.
// salt must be saltLen bytes; if nil, a new random salt is generated.
// Returns key and the salt used (needed for re-derivation on restart).
func DeriveKey(passphrase []byte, salt []byte) (key []byte, usedSalt []byte, err error) {
	if salt == nil {
		salt = make([]byte, saltLen)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, fmt.Errorf("salt generation: %w", err)
		}
	}
	if len(salt) != saltLen {
		return nil, nil, fmt.Errorf("salt must be %d bytes, got %d", saltLen, len(salt))
	}
	key = argon2.IDKey(passphrase, salt, argon2Time, argon2Memory, argon2Threads, keyLen)
	return key, salt, nil
}
