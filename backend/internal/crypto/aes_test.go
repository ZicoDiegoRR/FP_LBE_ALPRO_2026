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

func TestAESEdgeCasesAndBinary(t *testing.T) {
	key := []byte("kunci-rahasia-pmo-harus-32-byte!")
	cases := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"1_byte_zero", []byte{0x00}},
		{"binary_00_ff", []byte{0x00, 0xff, 0x01, 0x7f, 0x80}},
		{"15_bytes", bytes.Repeat([]byte("A"), 15)},
		{"16_bytes", bytes.Repeat([]byte("B"), 16)},
		{"17_bytes", bytes.Repeat([]byte("C"), 17)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := EncryptAES(tc.data, key)
			if err != nil {
				t.Fatalf("EncryptAES gagal (%s): %v", tc.name, err)
			}
			dec, err := DecryptAES(enc, key)
			if err != nil {
				t.Fatalf("DecryptAES gagal (%s): %v", tc.name, err)
			}
			if !bytes.Equal(dec, tc.data) {
				t.Fatalf("round-trip tidak cocok (%s)", tc.name)
			}
		})
	}
}

func TestAESNonDeterministicAndTampering(t *testing.T) {
	key := []byte("kunci-rahasia-pmo-harus-32-byte!")
	plain := []byte("data uji non-deterministik dan tampering")

	enc1, err := EncryptAES(plain, key)
	if err != nil {
		t.Fatalf("EncryptAES #1 gagal: %v", err)
	}
	enc2, err := EncryptAES(plain, key)
	if err != nil {
		t.Fatalf("EncryptAES #2 gagal: %v", err)
	}
	if bytes.Equal(enc1, enc2) {
		t.Fatal("dua enkripsi plaintext sama harus menghasilkan ciphertext berbeda karena nonce acak")
	}

	// Tamper 1 byte terakhir (auth tag / ciphertext)
	tampered := append([]byte(nil), enc1...)
	tampered[len(tampered)-1] ^= 0x01
	if _, err := DecryptAES(tampered, key); err == nil {
		t.Fatal("tampering 1 byte seharusnya ditolak oleh AES-GCM")
	}

	// Uji panjang kunci salah & ciphertext terpotong
	if _, err := EncryptAES(plain, []byte("short-key")); err == nil {
		t.Fatal("kunci tidak 32 byte seharusnya ditolak")
	}
	if _, err := DecryptAES(enc1[:5], key); err == nil {
		t.Fatal("ciphertext terpotong seharusnya ditolak")
	}
}