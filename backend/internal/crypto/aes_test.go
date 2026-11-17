package crypto

import (
	"bytes"
	"testing"
)

func TestAESEncryptionDecryption(t *testing.T) {
	// Kunci AES-256 harus tepat 32 byte (32 karakter ASCII = 32 byte)
	key := []byte("kunci-rahasia-pmo-harus-32-byte!")
	plaintextAsli := []byte("Halo Dosen Penguji, ini data peminjaman rahasia!")

	// 1. Uji Enkripsi
	ciphertext, err := EncryptAES(plaintextAsli, key)
	if err != nil {
		t.Fatalf("Enkripsi AES gagal: %v", err)
	}

	// Pastikan ciphertext tidak sama dengan plaintext asli
	if bytes.Equal(ciphertext, plaintextAsli) {
		t.Fatal("Ciphertext tidak boleh sama dengan plaintext!")
	}

	// 2. Uji Dekripsi dengan kunci yang benar
	hasilDekripsi, err := DecryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("Dekripsi AES gagal: %v", err)
	}

	// Pastikan hasil dekripsi identik 100% dengan teks asal
	if !bytes.Equal(hasilDekripsi, plaintextAsli) {
		t.Fatalf("Hasil dekripsi tidak cocok! Didapat: %s, Seharusnya: %s", string(hasilDekripsi), string(plaintextAsli))
	}

	// 3. Uji Keamanan: Dekripsi dengan KUNCI SALAH harus ditolak!
	kunciPalsu := []byte("kunci-yang-salah-tetap-32-byte!!")
	_, err = DecryptAES(ciphertext, kunciPalsu)
	if err == nil {
		t.Fatal("Dekripsi seharusnya gagal saat menggunakan kunci yang salah!")
	}

	t.Log("Uji coba AES-256-GCM Sukses 100%: Integritas & Kerahasiaan Terbukti!")
}