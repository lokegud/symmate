package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

var ErrDecryptFailed = errors.New("decryption failed: wrong key or corrupted ciphertext")

// Encrypt encrypts plaintext using XChaCha20-Poly1305.
// ad is additional data bound to the ciphertext but not encrypted (may be nil).
// Returns nonce || ciphertext.
func Encrypt(key []byte, plaintext, ad []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("aead init: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	ct := aead.Seal(nonce, nonce, plaintext, ad)
	return ct, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
// Returns ErrDecryptFailed on authentication error.
func Decrypt(key []byte, ciphertext, ad []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("aead init: %w", err)
	}
	ns := aead.NonceSize()
	if len(ciphertext) < ns {
		return nil, ErrDecryptFailed
	}
	nonce, ct := ciphertext[:ns], ciphertext[ns:]
	pt, err := aead.Open(nil, nonce, ct, ad)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return pt, nil
}
