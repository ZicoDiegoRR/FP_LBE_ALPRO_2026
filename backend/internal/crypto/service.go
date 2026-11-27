package crypto

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"pmo-backend/internal/utils"
)

// CryptoService adalah penyedia layanan enkripsi dan dekripsi beserta audit logging
type CryptoService struct{}

// NewCryptoService membuat instance baru CryptoService
func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

// Encrypt mendistribusikan proses enkripsi ke algoritma yang dipilih
func (s *CryptoService) Encrypt(data []byte, algorithm string, key []byte) ([]byte, error) {
	algo := strings.ToUpper(strings.TrimSpace(algorithm))
	start := time.Now()

	var (
		out []byte
		err error
	)
	switch algo {
	case "AES":
		out, err = EncryptAES(data, key)
	case "DES":
		out, err = EncryptDES(data, key)
	case "RC4":
		out, err = EncryptRC4(data, key)
	default:
		err = errors.New("algoritma tidak didukung: gunakan AES, DES, atau RC4")
	}

	utils.LogCryptoOperation("ENCRYPT", algo, len(data), time.Since(start), err)
	return out, err
}

// Decrypt mendistribusikan proses dekripsi ke algoritma yang dipilih
func (s *CryptoService) Decrypt(data []byte, algorithm string, key []byte) ([]byte, error) {
	algo := strings.ToUpper(strings.TrimSpace(algorithm))
	start := time.Now()

	var (
		out []byte
		err error
	)
	switch algo {
	case "AES":
		out, err = DecryptAES(data, key)
	case "DES":
		out, err = DecryptDES(data, key)
	case "RC4":
		out, err = DecryptRC4(data, key)
	default:
		err = errors.New("algoritma tidak didukung: gunakan AES, DES, atau RC4")
	}

	utils.LogCryptoOperation("DECRYPT", algo, len(data), time.Since(start), err)
	return out, err
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