package crypto

import (
	"crypto/rc4"
	"errors"
)

// EncryptRC4 mengenkripsi data menggunakan RC4 Stream Cipher
// Panjang kunci RC4 umumnya antara 1 sampai 256 byte (kita standarkan minimal 16 byte untuk keamanan)
func EncryptRC4(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) < 16 {
		return nil, errors.New("kunci RC4 minimal harus berukuran 16 byte (128 bit)")
	}

	// Inisialisasi state RC4 (Key Scheduling Algorithm / KSA)
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Alokasikan memori untuk ciphertext dengan ukuran yang persis sama dengan plaintext
	ciphertext := make([]byte, len(plaintext))

	// XOR plaintext dengan keystream yang dihasilkan PRGA (Pseudo-Random Generation Algorithm)
	cipher.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

// DecryptRC4 mendekripsi ciphertext RC4
// Sifat simetris XOR: Encrypt dan Decrypt menggunakan operasi matematika yang identik
func DecryptRC4(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) < 16 {
		return nil, errors.New("kunci RC4 minimal harus berukuran 16 byte (128 bit)")
	}

	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	// XOR kembali ciphertext dengan keystream yang sama untuk mendapatkan plaintext
	cipher.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}