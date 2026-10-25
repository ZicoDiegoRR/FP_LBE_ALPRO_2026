package rc4_encrypt

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestRC4CryptRoundTrip(t *testing.T) {
	// Data pengujian
	key := []byte("testing-key")
	plaintext := []byte("Hello World")

	// Enkripsi plaintext
	ciphertext, err := RC4Crypt(plaintext, key)
	if err != nil {
		t.Fatalf("Enkripsi gagal: %v", err)
	}

	// Dekripsi ciphertext menggunakan kunci yang sama
	decrypted, err := RC4Crypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Dekripsi gagal: %v", err)
	}

	// Pastikan hasil dekripsi sama dengan plaintext
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatal("Hasil dekripsi tidak sama dengan plaintext")
	}

	// Tampilkan hasil pengujian di terminal
	t.Logf("Plaintext  : %s", plaintext)
	t.Logf("Ciphertext : %s", hex.EncodeToString(ciphertext))
	t.Logf("Decrypted  : %s", decrypted)
}
