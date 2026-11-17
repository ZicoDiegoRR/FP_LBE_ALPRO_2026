package crypto

import (
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