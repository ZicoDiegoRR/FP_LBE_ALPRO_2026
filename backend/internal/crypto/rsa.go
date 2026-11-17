package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// GenerateRSAKeyPair menghasilkan pasangan kunci RSA 2048-bit dalam format PEM string
func GenerateRSAKeyPair() (string, string, error) {
	// 1. Generate Private Key 2048-bit
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// 2. Encode Private Key ke format PEM
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	// 3. Encode Public Key ke format PEM
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(pubPEM), string(privPEM), nil
}

// EncryptRSAOAEP mengenkripsi data kecil (misal kunci simetris AES 32-byte) dengan Public Key penerima
func EncryptRSAOAEP(plaintext []byte, publicKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("gagal membaca format PEM public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("bukan tipe RSA public key")
	}

	// Gunakan RSA-OAEP dengan SHA-256 untuk keamanan terhadap padding oracle attack
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, plaintext, nil)
}

// DecryptRSAOAEP mendekripsi ciphertext menggunakan Private Key pemilik
func DecryptRSAOAEP(ciphertext []byte, privateKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("gagal membaca format PEM private key")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, ciphertext, nil)
}

// SignDocument membuat Tanda Tangan Digital atas dokumen menggunakan RSA-PSS / PKCS1v15
func SignDocument(fileBytes []byte, privateKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("gagal membaca format PEM private key")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 1. Hitung Hash SHA-256 dokumen
	hashed := sha256.Sum256(fileBytes)

	// 2. Tandatangani hash dokumen dengan private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// VerifyDocumentSignature memverifikasi apakah tanda tangan digital valid dan dokumen belum dimanipulasi
func VerifyDocumentSignature(fileBytes []byte, signature []byte, publicKeyPEM string) error {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return errors.New("gagal membaca format PEM public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return errors.New("bukan tipe RSA public key")
	}

	// 1. Hitung Hash dokumen saat ini
	hashed := sha256.Sum256(fileBytes)

	// 2. Verifikasi tanda tangan dengan public key
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature)
}