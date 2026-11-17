package crypto

import (
	"bytes"
	"testing"
)

func TestDESEncryptionDecryption(t *testing.T) {
	// Kunci DES harus tepat 8 byte (8 karakter ASCII = 8 byte = 64 bit)
	key := []byte("pmo-key8")
	plaintextAsli := []byte("Dokumen rahasia peminjaman dengan DES!")

	// 1. Uji Enkripsi
	ciphertext, err := EncryptDES(plaintextAsli, key)
	if err != nil {
		t.Fatalf("Enkripsi DES gagal: %v", err)
	}

	// Pastikan ciphertext tidak sama dengan plaintext asli
	if bytes.Equal(ciphertext, plaintextAsli) {
		t.Fatal("Ciphertext tidak boleh sama dengan plaintext!")
	}

	// 2. Uji Dekripsi dengan kunci yang benar
	hasilDekripsi, err := DecryptDES(ciphertext, key)
	if err != nil {
		t.Fatalf("Dekripsi DES gagal: %v", err)
	}

	// Pastikan hasil dekripsi identik dengan teks asal
	if !bytes.Equal(hasilDekripsi, plaintextAsli) {
		t.Fatalf("Hasil dekripsi tidak cocok! Didapat: %s, Seharusnya: %s", string(hasilDekripsi), string(plaintextAsli))
	}

	// 3. Uji Keamanan: Dekripsi dengan KUNCI SALAH harus menghasilkan error atau data rusak
	kunciPalsu := []byte("salahkey")
	_, err = DecryptDES(ciphertext, kunciPalsu)
	if err == nil {
		t.Fatal("Dekripsi DES dengan kunci salah seharusnya gagal unpadding!")
	}

	t.Log("Uji coba DES-CBC Sukses 100%: Padding PKCS#7 & Enkripsi Berhasil!")
}