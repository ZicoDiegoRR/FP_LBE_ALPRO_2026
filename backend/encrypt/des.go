package encrypt

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Buat padding apabila melebihi 7
func padDES(b []byte) []byte {
	n := des.BlockSize - len(b)%des.BlockSize
	return append(b, bytes.Repeat([]byte{byte(n)}, n)...)
}

func unpadDES(b []byte) ([]byte, error) {
	if len(b) == 0 || len(b)%des.BlockSize != 0 {
		return nil, errors.New("bad length")
	}
	n := int(b[len(b)-1])
	if n == 0 || n > des.BlockSize || !bytes.Equal(b[len(b)-n:], bytes.Repeat([]byte{byte(n)}, n)) {
		return nil, errors.New("bad padding")
	}
	return b[:len(b)-n], nil
}

func EncryptDES(key, plaintext []byte) ([]byte, error) {
	block, err := des.NewCipher(key) // key must be exactly 8 bytes
	if err != nil {
		return nil, err
	}
	p := padDES(plaintext)
	out := make([]byte, des.BlockSize+len(p))
	iv := out[:des.BlockSize]
	if _, err := rand.Read(iv); err != nil { // fresh random IV per message
		return nil, err
	}
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out[des.BlockSize:], p)
	return out, nil // IV || ciphertext
}

func DecryptDES(key, data []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) < 2*des.BlockSize ||
		(len(data)-des.BlockSize)%des.BlockSize != 0 {
		return nil, errors.New("invalid ciphertext length")
	}
	iv, ct := data[:des.BlockSize], data[des.BlockSize:]
	out := make([]byte, len(ct))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, ct)
	return unpadDES(out)
}

// EncryptFileDES reads inputPath, encrypts it with EncryptDES, and writes
// the result to inputPath+".enc" (or outputPath, if given). Returns the
// path written.
func EncryptFileDES(key []byte, inputPath string, outputPath ...string) (string, error) {
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("reading input file: %w", err)
	}

	ciphertext, err := EncryptDES(key, plaintext)
	if err != nil {
		return "", fmt.Errorf("encrypting: %w", err)
	}

	outPath := inputPath + ".enc"
	if len(outputPath) > 0 && outputPath[0] != "" {
		outPath = outputPath[0]
	}

	if err := os.WriteFile(outPath, ciphertext, 0600); err != nil {
		return "", fmt.Errorf("writing encrypted file: %w", err)
	}

	return outPath, nil
}

// DecryptFileDES is the counterpart to EncryptFileDES.
func DecryptFileDES(key []byte, inputPath string, outputPath ...string) (string, error) {
	ciphertext, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("reading input file: %w", err)
	}

	plaintext, err := DecryptDES(key, ciphertext)
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

	if err := os.WriteFile(outPath, plaintext, 0600); err != nil {
		return "", fmt.Errorf("writing decrypted file: %w", err)
	}

	return outPath, nil
}