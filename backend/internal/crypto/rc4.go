package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rc4"
	"crypto/sha256"
	"errors"
	"io"
)

const rc4SaltSize = 16

func deriveRC4Subkey(key, salt []byte, purpose string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(purpose))
	mac.Write(salt)
	return mac.Sum(nil)
}

// EncryptRC4 mengenkripsi data dengan salt acak dan MAC.
// Format hasil: [salt 16 byte][ciphertext][HMAC-SHA256].
func EncryptRC4(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) < 16 {
		return nil, errors.New("kunci RC4 minimal harus berukuran 16 byte (128 bit)")
	}

	ciphertext := make([]byte, rc4SaltSize+len(plaintext)+sha256.Size)
	salt := ciphertext[:rc4SaltSize]
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	encryptionKey := deriveRC4Subkey(key, salt, "RC4 encryption key")
	stream, err := rc4.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	authenticatedData := ciphertext[:len(ciphertext)-sha256.Size]
	encryptedData := authenticatedData[rc4SaltSize:]
	stream.XORKeyStream(encryptedData, plaintext)

	macKey := deriveRC4Subkey(key, salt, "RC4 authentication key")
	mac := hmac.New(sha256.New, macKey)
	mac.Write(authenticatedData)
	copy(ciphertext[len(authenticatedData):], mac.Sum(nil))

	return ciphertext, nil
}

// DecryptRC4 memverifikasi MAC sebelum mendekripsi ciphertext.
func DecryptRC4(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) < 16 {
		return nil, errors.New("kunci RC4 minimal harus berukuran 16 byte (128 bit)")
	}
	if len(ciphertext) < rc4SaltSize+sha256.Size {
		return nil, errors.New("ciphertext RC4 terlalu pendek atau tidak memiliki salt dan MAC")
	}

	authenticatedData := ciphertext[:len(ciphertext)-sha256.Size]
	salt := authenticatedData[:rc4SaltSize]
	macKey := deriveRC4Subkey(key, salt, "RC4 authentication key")
	mac := hmac.New(sha256.New, macKey)
	mac.Write(authenticatedData)
	if !hmac.Equal(ciphertext[len(authenticatedData):], mac.Sum(nil)) {
		return nil, errors.New("integritas ciphertext RC4 tidak valid")
	}

	encryptionKey := deriveRC4Subkey(key, salt, "RC4 encryption key")
	stream, err := rc4.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	encryptedData := authenticatedData[rc4SaltSize:]
	plaintext := make([]byte, len(encryptedData))
	stream.XORKeyStream(plaintext, encryptedData)
	return plaintext, nil
}
