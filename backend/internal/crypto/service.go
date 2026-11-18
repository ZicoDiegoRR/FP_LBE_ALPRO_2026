package crypto

import (
	"encoding/base64"
	"errors"
	"strings"
)

// CryptoService adalah interface penyedia layanan enkripsi dan dekripsi
type CryptoService struct{}

// NewCryptoService membuat instance baru CryptoService
func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

// Encrypt mendistribusikan proses enkripsi ke algoritma yang dipilih
func (s *CryptoService) Encrypt(data []byte, algorithm string, key []byte) ([]byte, error) {
	switch strings.ToUpper(algorithm) {
	case "AES":
		return EncryptAES(data, key)
	case "DES":
		return EncryptDES(data, key)
	case "RC4":
		return EncryptRC4(data, key)
	default:
		return nil, errors.New("algoritma tidak didukung: gunakan AES, DES, atau RC4")
	}
}

// Decrypt mendistribusikan proses dekripsi ke algoritma yang dipilih
func (s *CryptoService) Decrypt(data []byte, algorithm string, key []byte) ([]byte, error) {
	switch strings.ToUpper(algorithm) {
	case "AES":
		return DecryptAES(data, key)
	case "DES":
		return DecryptDES(data, key)
	case "RC4":
		return DecryptRC4(data, key)
	default:
		return nil, errors.New("algoritma tidak didukung: gunakan AES, DES, atau RC4")
	}
}

// EncryptString mengenkripsi string plaintext dan mengembalikan hasil terenkripsi dalam format Base64
func (s *CryptoService) EncryptString(plaintext string, algorithm string, key []byte) (string, error) {
	cipherBytes, err := s.Encrypt([]byte(plaintext), algorithm, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

// DecryptString mendekripsi string Base64 hasil EncryptString menjadi plaintext asli
func (s *CryptoService) DecryptString(ciphertextBase64 string, algorithm string, key []byte) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", errors.New("format ciphertext Base64 tidak valid")
	}
	plainBytes, err := s.Decrypt(cipherBytes, algorithm, key)
	if err != nil {
		return "", err
	}
	return string(plainBytes), nil
}