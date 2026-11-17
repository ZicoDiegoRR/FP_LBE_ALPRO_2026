package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptAES mengenkripsi plaintext menggunakan AES-256-GCM
// Membutuhkan key sepanjang 32 byte (256 bit)
func EncryptAES(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("kunci AES harus berukuran 32 byte untuk AES-256")
	}

	// 1. Inisialisasi cipher block AES
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. Bungkus block cipher ke dalam mode GCM (AEAD)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 3. Buat Nonce (Number used once) acak berukuran standar GCM (12 byte)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 4. Enkripsi dan tempelkan ciphertext setelah nonce
	// Format hasil: [NONCE 12 byte][CIPHERTEXT + AUTH_TAG 16 byte]
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// DecryptAES mendekripsi ciphertext hasil EncryptAES menggunakan key yang sama
func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("kunci AES harus berukuran 32 byte untuk AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ukuran ciphertext terlalu pendek (tidak memiliki nonce)")
	}

	// Pisahkan nonce (12 byte pertama) dari ciphertext sebenarnya
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Dekripsi dan verifikasi keaslian data menggunakan tag autentikasi bawaan GCM
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, errors.New("dekripsi gagal: kunci salah atau data telah dimanipulasi")
	}

	return plaintext, nil
}