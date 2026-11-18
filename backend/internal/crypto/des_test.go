package crypto

import (
	"bytes"
	"crypto/des"
	"crypto/sha256"
	"strconv"
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
		t.Fatal("Dekripsi DES dengan kunci salah seharusnya gagal memverifikasi MAC!")
	}
}

func TestDESBlockBoundaryRoundTrips(t *testing.T) {
	key := []byte("pmo-key8")
	for _, size := range []int{0, 1, 7, 8, 9, 15, 16, 17} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			plaintext := make([]byte, size)
			for i := range plaintext {
				plaintext[i] = byte(i)
			}

			ciphertext, err := EncryptDES(plaintext, key)
			if err != nil {
				t.Fatalf("Enkripsi DES untuk %d byte gagal: %v", size, err)
			}

			expectedSize := des.BlockSize + (size/des.BlockSize+1)*des.BlockSize + sha256.Size
			if len(ciphertext) != expectedSize {
				t.Fatalf("Ukuran ciphertext = %d, ingin %d", len(ciphertext), expectedSize)
			}

			decrypted, err := DecryptDES(ciphertext, key)
			if err != nil {
				t.Fatalf("Dekripsi DES untuk %d byte gagal: %v", size, err)
			}
			if !bytes.Equal(decrypted, plaintext) {
				t.Fatalf("Hasil dekripsi untuk %d byte tidak cocok", size)
			}
		})
	}
}

func TestDESRejectsTruncatedCiphertext(t *testing.T) {
	key := []byte("pmo-key8")
	ciphertext, err := EncryptDES([]byte("plaintext untuk uji pemotongan"), key)
	if err != nil {
		t.Fatalf("Enkripsi DES gagal: %v", err)
	}

	truncated := [][]byte{
		nil,
		ciphertext[:7],
		ciphertext[:8],
		ciphertext[:len(ciphertext)-1],
		ciphertext[:len(ciphertext)-32],
	}
	for i, input := range truncated {
		if _, err := DecryptDES(input, key); err == nil {
			t.Errorf("Ciphertext terpotong ke-%d seharusnya ditolak", i)
		}
	}
}

func TestDESRejectsTamperedCiphertext(t *testing.T) {
	key := []byte("pmo-key8")
	ciphertext, err := EncryptDES([]byte("plaintext untuk uji integritas"), key)
	if err != nil {
		t.Fatalf("Enkripsi DES gagal: %v", err)
	}

	for _, index := range []int{0, 8, len(ciphertext) - 1} {
		tampered := append([]byte(nil), ciphertext...)
		tampered[index] ^= 0x01
		if _, err := DecryptDES(tampered, key); err == nil {
			t.Errorf("Ciphertext dengan byte ke-%d diubah seharusnya ditolak", index)
		}
	}
}
