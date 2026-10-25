// AES 256
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func EncryptAES(plainText, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherText := gcm.Seal(nonce, nonce, plainText, nil)
	return cipherText, nil
}

func DecryptAES(cipherText, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
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
	if len(cipherText) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, actualCipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// EncryptFileAES reads inputPath, encrypts it with EncryptAES, and writes
// the result to inputPath+".enc" (or outputPath, if given). Returns the
// path written.
func EncryptFileAES(inputPath string, key []byte, outputPath ...string) (string, error) {
	plainText, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("reading input file: %w", err)
	}

	cipherText, err := EncryptAES(plainText, key)
	if err != nil {
		return "", fmt.Errorf("encrypting: %w", err)
	}

	outPath := inputPath + ".enc"
	if len(outputPath) > 0 && outputPath[0] != "" {
		outPath = outputPath[0]
	}

	if err := os.WriteFile(outPath, cipherText, 0600); err != nil {
		return "", fmt.Errorf("writing encrypted file: %w", err)
	}

	return outPath, nil
}

// DecryptFileAES is the counterpart to EncryptFileAES.
func DecryptFileAES(inputPath string, key []byte, outputPath ...string) (string, error) {
	cipherText, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("reading input file: %w", err)
	}

	plainText, err := DecryptAES(cipherText, key)
	if err != nil {
		return "", fmt.Errorf("decrypting: %w", err)
	}

	outPath := inputPath + ".dec"
	if ext := filepath.Ext(inputPath); ext == ".enc" {
		outPath = inputPath[:len(inputPath)-len(ext)]
	}
	if len(outputPath) > 0 && outputPath[0] != "" {
		outPath = outputPath[0]
	}

	if err := os.WriteFile(outPath, plainText, 0600); err != nil {
		return "", fmt.Errorf("writing decrypted file: %w", err)
	}

	return outPath, nil
}