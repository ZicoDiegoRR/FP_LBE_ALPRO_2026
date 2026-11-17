
package crypto

import (
	"bytes"
	"testing"
)

func TestRC4EncryptionDecryption(t *testing.T) {
	// Kunci RC4 minimal 16 byte (128 bit)
	key := []byte("pmo-key-rahasia-16")
	plaintextAsli := []byte("Dokumen rahasia peminjaman dengan RC4!")

	// 1. Uji Enkripsi
	ciphertext, err := EncryptRC4(plaintextAsli, key)
	if err != nil {
		t.Fatalf("Enkripsi RC4 gagal: %v", err)
	}

	if bytes.Equal(ciphertext, plaintextAsli) {
		t.Fatal("Ciphertext tidak boleh sama dengan plaintext!")
	}

	// 2. Uji Dekripsi dengan kunci yang benar
	hasilDekripsi, err := DecryptRC4(ciphertext, key)
	if err != nil {
		t.Fatalf("Dekripsi RC4 gagal: %v", err)
	}

	if !bytes.Equal(hasilDekripsi, plaintextAsli) {
		t.Fatalf(
			"Hasil dekripsi tidak cocok! Didapat: %s, Seharusnya: %s",
			string(hasilDekripsi),
			string(plaintextAsli),
		)
	}

	// 3. Uji Dekripsi dengan kunci yang salah
	kunciPalsu := []byte("kunci-yang-berbeda-16")

	hasilSalah, err := DecryptRC4(ciphertext, kunciPalsu)
	if err != nil {
		t.Fatalf("Dekripsi RC4 gagal: %v", err)
	}

	// RC4 tidak memiliki autentikasi.
	// Kunci salah menghasilkan data berbeda tanpa error.
	if bytes.Equal(hasilSalah, plaintextAsli) {
		t.Fatal("Dekripsi dengan kunci salah menghasilkan plaintext asli!")
	}

	t.Log("Uji coba RC4 berhasil: Enkripsi dan dekripsi berjalan!")
}
