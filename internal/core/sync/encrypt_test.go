package sync

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	plaintext := []byte(`{"version":1,"commands":[]}`)
	passphrase := "my-secret-key-123"

	encoded, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := Decrypt(encoded, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(plaintext, decoded) {
		t.Fatalf("plaintext = %q, decoded = %q", plaintext, decoded)
	}
}

func TestEncryptEmpty(t *testing.T) {
	encoded, err := Encrypt([]byte{}, "key")
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) == 0 {
		t.Fatal("encoded output should not be empty")
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	plaintext := []byte("secret data")
	encoded, err := Encrypt(plaintext, "correct-key")
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decrypt(encoded, "wrong-key")
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	_, err := Decrypt([]byte("not-valid-base64!!!"), "key")
	if err == nil {
		t.Fatal("expected error for corrupted data")
	}

	// valid base64 but not valid ciphertext
	_, err = Decrypt([]byte("AAAAAA=="), "key")
	if err == nil {
		t.Fatal("expected error for invalid ciphertext")
	}
}

func TestDeterministicWithDifferentNonces(t *testing.T) {
	e1, _ := Encrypt([]byte("same data"), "key")
	e2, _ := Encrypt([]byte("same data"), "key")

	if bytes.Equal(e1, e2) {
		t.Fatal("encrypted output should differ due to random nonce")
	}
}

func TestLongPassphrase(t *testing.T) {
	plaintext := []byte("hello")
	encoded, err := Encrypt(plaintext, "a-very-long-passphrase-that-exceeds-thirty-two-bytes-in-length-for-testing")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decrypt(encoded, "a-very-long-passphrase-that-exceeds-thirty-two-bytes-in-length-for-testing")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, decoded) {
		t.Fatal("long passphrase roundtrip failed")
	}
}
