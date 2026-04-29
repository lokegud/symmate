package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, _, err := DeriveKey([]byte("test-passphrase"), nil)
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("hello, symmate")
	ad := []byte("conversation:123")

	ct, err := Encrypt(key, plaintext, ad)
	if err != nil {
		t.Fatal(err)
	}
	pt, err := Decrypt(key, ct, ad)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Fatalf("plaintext mismatch: got %q want %q", pt, plaintext)
	}
}

func TestWrongKeyFails(t *testing.T) {
	key, _, err := DeriveKey([]byte("correct-passphrase"), nil)
	if err != nil {
		t.Fatal(err)
	}
	wrongKey, _, err := DeriveKey([]byte("wrong-passphrase"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ct, err := Encrypt(key, []byte("secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Decrypt(wrongKey, ct, nil)
	if err != ErrDecryptFailed {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestWrongADFails(t *testing.T) {
	key, _, err := DeriveKey([]byte("passphrase"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ct, err := Encrypt(key, []byte("secret"), []byte("good-ad"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Decrypt(key, ct, []byte("bad-ad"))
	if err != ErrDecryptFailed {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestDeriveKeySameSaltSameKey(t *testing.T) {
	pass := []byte("reproducible")
	k1, salt, err := DeriveKey(pass, nil)
	if err != nil {
		t.Fatal(err)
	}
	k2, _, err := DeriveKey(pass, salt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(k1, k2) {
		t.Fatal("same passphrase+salt must produce same key")
	}
}

func TestNilADAccepted(t *testing.T) {
	key, _, err := DeriveKey([]byte("p"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ct, err := Encrypt(key, []byte("data"), nil)
	if err != nil {
		t.Fatal(err)
	}
	pt, err := Decrypt(key, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(pt) != "data" {
		t.Fatal("round-trip failed with nil AD")
	}
}
