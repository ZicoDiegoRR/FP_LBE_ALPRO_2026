package crypto

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"testing"
)

func TestRC4EncryptionDecryption(t *testing.T) {
	key := []byte("pmo-key-rahasia-16")
	plaintext := []byte("Dokumen rahasia peminjaman dengan RC4!")

	ciphertext, err := EncryptRC4(plaintext, key)
	if err != nil {
		t.Fatalf("Enkripsi RC4 gagal: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("Ciphertext tidak boleh sama dengan plaintext")
	}

	decrypted, err := DecryptRC4(ciphertext, key)
	if err != nil {
		t.Fatalf("Dekripsi RC4 gagal: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("Hasil dekripsi tidak cocok: didapat %q, ingin %q", decrypted, plaintext)
	}

	secondCiphertext, err := EncryptRC4(plaintext, key)
	if err != nil {
		t.Fatalf("Enkripsi RC4 kedua gagal: %v", err)
	}
	if bytes.Equal(ciphertext, secondCiphertext) {
		t.Fatal("Enkripsi berulang dengan plaintext dan kunci sama seharusnya menghasilkan salt acak berbeda")
	}

	if _, err := DecryptRC4(ciphertext, []byte("kunci-yang-berbeda-16")); err == nil {
		t.Fatal("Dekripsi dengan kunci yang salah seharusnya gagal memverifikasi MAC")
	}
}

func TestRC4BoundaryRoundTrips(t *testing.T) {
	key := []byte("pmo-key-rahasia-16")
	for _, size := range []int{0, 1, 15, 16, 17, 31, 32, 33} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			plaintext := make([]byte, size)
			for i := range plaintext {
				plaintext[i] = byte(i)
			}
			if size > 1 {
				plaintext[1] = 0xff
			}

			ciphertext, err := EncryptRC4(plaintext, key)
			if err != nil {
				t.Fatalf("Enkripsi RC4 untuk %d byte gagal: %v", size, err)
			}
			expectedSize := rc4SaltSize + size + sha256.Size
			if len(ciphertext) != expectedSize {
				t.Fatalf("Ukuran ciphertext = %d, ingin %d", len(ciphertext), expectedSize)
			}

			decrypted, err := DecryptRC4(ciphertext, key)
			if err != nil {
				t.Fatalf("Dekripsi RC4 untuk %d byte gagal: %v", size, err)
			}
			if !bytes.Equal(decrypted, plaintext) {
				t.Fatalf("Hasil dekripsi untuk %d byte tidak cocok", size)
			}
		})
	}
}

func TestRC4RejectsTruncatedCiphertext(t *testing.T) {
	key := []byte("pmo-key-rahasia-16")
	ciphertext, err := EncryptRC4([]byte("plaintext untuk uji pemotongan minimal 32 byte"), key)
	if err != nil {
		t.Fatalf("Enkripsi RC4 gagal: %v", err)
	}

	truncated := [][]byte{
		nil,
		ciphertext[:rc4SaltSize-1],
		ciphertext[:rc4SaltSize],
		ciphertext[:len(ciphertext)-1],
		ciphertext[:len(ciphertext)-sha256.Size],
	}
	for i, input := range truncated {
		if _, err := DecryptRC4(input, key); err == nil {
			t.Errorf("Ciphertext terpotong ke-%d seharusnya ditolak", i)
		}
	}
}

func TestRC4RejectsTamperedCiphertext(t *testing.T) {
	key := []byte("pmo-key-rahasia-16")
	ciphertext, err := EncryptRC4([]byte("plaintext untuk uji integritas"), key)
	if err != nil {
		t.Fatalf("Enkripsi RC4 gagal: %v", err)
	}

	for _, index := range []int{0, rc4SaltSize, len(ciphertext) - 1} {
		tampered := append([]byte(nil), ciphertext...)
		tampered[index] ^= 0x01
		if _, err := DecryptRC4(tampered, key); err == nil {
			t.Errorf("Ciphertext dengan byte ke-%d diubah seharusnya ditolak", index)
		}
	}
}
