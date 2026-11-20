package crypto

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// pkcs7Padding menambahkan bantalan agar panjang data menjadi kelipatan blockSize
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpadding menghapus bantalan setelah proses dekripsi
func pkcs7Unpadding(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("data terenkripsi tidak valid atau tidak sesuai ukuran blok")
	}

	padding := int(data[length-1])
	if padding > blockSize || padding == 0 {
		return nil, errors.New("padding tidak valid")
	}

	// Verifikasi bahwa seluruh byte padding bernilai sama
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("padding rusak atau kunci salah")
		}
	}

	return data[:length-padding], nil
}

// EncryptDES mengenkripsi data menggunakan DES mode CBC
// Membutuhkan key tepat 8 byte (64 bit)
func EncryptDES(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("kunci DES harus berukuran tepat 8 byte (64 bit)")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Terapkan padding PKCS#7 agar kelipatan 8 byte
	paddedPlaintext := pkcs7Padding(plaintext, des.BlockSize)

	// Buat IV (Initialization Vector) acak berukuran 8 byte
	ciphertext := make([]byte, des.BlockSize+len(paddedPlaintext))
	iv := ciphertext[:des.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Enkripsi menggunakan mode CBC
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[des.BlockSize:], paddedPlaintext)

	// Format hasil: [IV][ciphertext CBC][HMAC-SHA256].
	mac := hmac.New(sha256.New, key)
	mac.Write(ciphertext)

	authenticatedCiphertext := make([]byte, len(ciphertext)+sha256.Size)
	copy(authenticatedCiphertext, ciphertext)
	copy(authenticatedCiphertext[len(ciphertext):], mac.Sum(nil))
	return authenticatedCiphertext, nil
}

// DecryptDES mendekripsi data DES mode CBC
func DecryptDES(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("kunci DES harus berukuran tepat 8 byte (64 bit)")
	}

	if len(ciphertext) < des.BlockSize*2+sha256.Size {
		return nil, errors.New("ukuran ciphertext terlalu pendek atau tidak memiliki blok data dan MAC")
	}

	authenticatedData := ciphertext[:len(ciphertext)-sha256.Size]
	actualCiphertext := authenticatedData[des.BlockSize:]
	if len(actualCiphertext)%des.BlockSize != 0 {
		return nil, errors.New("panjang ciphertext bukan kelipatan ukuran blok DES")
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(authenticatedData)
	if !hmac.Equal(ciphertext[len(authenticatedData):], mac.Sum(nil)) {
		return nil, errors.New("integritas ciphertext DES tidak valid")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Lakukan dekripsi CBC
	mode := cipher.NewCBCDecrypter(block, authenticatedData[:des.BlockSize])
	decrypted := make([]byte, len(actualCiphertext))
	mode.CryptBlocks(decrypted, actualCiphertext)

	// Buang padding PKCS#7 untuk mendapatkan data asli
	plaintext, err := pkcs7Unpadding(decrypted, des.BlockSize)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
