package cryptox

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("hello, world! this is a test message.")

	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	got, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Fatalf("roundtrip failed: got %q, want %q", got, plaintext)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("secret")

	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	wrongKey := make([]byte, 32)
	wrongKey[0] = 1
	_, err = Decrypt(wrongKey, ciphertext)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestDecryptCiphertextTooShort(t *testing.T) {
	key := make([]byte, 32)
	_, err := Decrypt(key, []byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for short ciphertext")
	}
}

func TestEncryptToStringDecryptStringRoundtrip(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("test filename")

	encoded, err := EncryptToString(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptToString: %v", err)
	}

	got, err := DecryptString(key, encoded)
	if err != nil {
		t.Fatalf("DecryptString: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Fatalf("roundtrip failed: got %q, want %q", got, plaintext)
	}
}

func TestEncryptEmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)
	ciphertext, err := Encrypt(key, []byte{})
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}

	got, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected empty, got %d bytes", len(got))
	}
}
