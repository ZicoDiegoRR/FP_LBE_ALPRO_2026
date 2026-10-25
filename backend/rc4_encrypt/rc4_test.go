package rc4_encrypt

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRC4CryptRoundTrip(t *testing.T) {
	key := []byte("testing-key")
	original := []byte("Contoh data untuk pengujian RC4")

	encrypted, err := RC4Crypt(original, key)
	if err != nil {
		t.Fatalf("enkripsi gagal: %v", err)
	}

	if bytes.Equal(original, encrypted) {
		t.Fatal("hasil enkripsi sama dengan data asli")
	}

	decrypted, err := RC4Crypt(encrypted, key)
	if err != nil {
		t.Fatalf("dekripsi gagal: %v", err)
	}

	if !bytes.Equal(original, decrypted) {
		t.Fatal("hasil dekripsi tidak sama dengan data asli")
	}
}

func TestRC4FileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	key := []byte("testing-key")
	original := []byte("Isi file untuk pengujian")

	input := filepath.Join(dir, "input.txt")
	encrypted := filepath.Join(dir, "encrypted.bin")
	decrypted := filepath.Join(dir, "decrypted.txt")

	if err := os.WriteFile(input, original, 0600); err != nil {
		t.Fatal(err)
	}

	if err := RC4EncryptFile(input, encrypted, key); err != nil {
		t.Fatalf("enkripsi file gagal: %v", err)
	}

	if err := RC4DecryptFile(encrypted, decrypted, key); err != nil {
		t.Fatalf("dekripsi file gagal: %v", err)
	}

	result, err := os.ReadFile(decrypted)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(original, result) {
		t.Fatal("hasil file dekripsi tidak sama dengan file asli")
	}
}

func TestRC4RejectsEmptyKey(t *testing.T) {
	_, err := RC4Crypt([]byte("data"), nil)
	if err == nil {
		t.Fatal("seharusnya kunci kosong menghasilkan error")
	}
}
