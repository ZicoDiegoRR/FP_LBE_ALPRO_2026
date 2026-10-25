package rc4_encrypt

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestRC4CryptRoundTrip(t *testing.T) {
	key := []byte("testing-key")
	plaintext := []byte("Hello World")

	// Enkripsi
	ciphertext, err := RC4Crypt(plaintext, key)
	if err != nil {
		t.Fatalf("Enkripsi gagal: %v", err)
	}

	// Dekripsi menggunakan fungsi yang sama
	decrypted, err := RC4Crypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Dekripsi gagal: %v", err)
	}

	// Pastikan plaintext kembali utuh
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf(
			"Hasil dekripsi tidak sesuai: got %q, want %q",
			decrypted,
			plaintext,
		)
	}

	t.Logf("Plaintext  : %s", plaintext)
	t.Logf("Ciphertext : %s", hex.EncodeToString(ciphertext))
	t.Logf("Decrypted  : %s", decrypted)
}
