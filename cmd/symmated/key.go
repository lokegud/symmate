package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	symcrypto "github.com/loke/symmate/internal/crypto"
	"github.com/rs/zerolog/log"
)

// loadOrDeriveKey returns the 32-byte AEAD master key for hostDir.
//
// If SYMMATE_KEY_PASS is set: derives via Argon2id. The 16-byte salt is
// stored in <hostDir>/keychain.salt on first boot and reused on restart.
//
// Otherwise: loads or generates a random dev key in $XDG_CONFIG_HOME/symmated/dev.key.
// This is NOT suitable for production — data survives restart, but the key is unprotected.
func loadOrDeriveKey(hostDir string) ([]byte, error) {
	if pass := os.Getenv("SYMMATE_KEY_PASS"); pass != "" {
		return deriveFromPassphrase([]byte(pass), hostDir)
	}
	return loadOrGenerateDevKey()
}

func deriveFromPassphrase(passphrase []byte, hostDir string) ([]byte, error) {
	saltPath := filepath.Join(hostDir, "keychain.salt")

	var salt []byte
	if data, err := os.ReadFile(saltPath); err == nil && len(data) == 16 {
		salt = data
	}

	key, usedSalt, err := symcrypto.DeriveKey(passphrase, salt)
	if err != nil {
		return nil, fmt.Errorf("key derivation: %w", err)
	}

	if salt == nil {
		if err := os.WriteFile(saltPath, usedSalt, 0600); err != nil {
			return nil, fmt.Errorf("write keychain.salt: %w", err)
		}
		log.Info().Str("path", saltPath).Msg("keychain salt written")
	}
	return key, nil
}

func loadOrGenerateDevKey() ([]byte, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	keyDir := filepath.Join(configDir, "symmated")
	keyPath := filepath.Join(keyDir, "dev.key")

	if data, err := os.ReadFile(keyPath); err == nil && len(data) == 32 {
		log.Warn().Str("path", keyPath).Msg("using dev key — set SYMMATE_KEY_PASS for production")
		return data, nil
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate dev key: %w", err)
	}
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, fmt.Errorf("create key dir: %w", err)
	}
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("write dev key: %w", err)
	}
	log.Warn().Str("path", keyPath).Msg("generated dev key — set SYMMATE_KEY_PASS for production")
	return key, nil
}
