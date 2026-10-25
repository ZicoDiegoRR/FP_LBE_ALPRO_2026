package encrypt

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"
	"pmo-backend/utils"
)

type Algorithm string

const (
	AES Algorithm = "aes"
	DES Algorithm = "des"
	RC4 Algorithm = "rc4"
)

func GetEncryptionKey() []byte {
	if key := os.Getenv("ENCRYPTION_SECRET_KEY"); key != "" {
		return []byte(key)
	}
	// Fallback key (must be 16, 24, or 32 bytes for AES)
	return []byte("dev-only-secret-key-32-bytes-long!")
}

func GetAlgorithm() Algorithm {
	if algo := os.Getenv("ENCRYPTION_ALGO"); algo != "" {
		return Algorithm(algo)
	}
	return AES
}

// Middleware encrypts input with automatic file/string detection and logging.
func Middleware(input string, key []byte, algo Algorithm) (string, error) {
	start := time.Now()
	inputType := "string"
	if info, err := os.Stat(input); err == nil && !info.IsDir() {
		inputType = "file"
	}

	utils.Log(utils.LevelInfo, "Starting encryption | Algorithm: %s | Type: %s | Target: %s", algo, inputType, input)

	var output string
	var err error

	if inputType == "file" {
		output, err = EncryptFilePath(input, key, algo)
	} else {
		output, err = EncryptString(input, key, algo)
	}

	latency := time.Since(start)

	if err != nil {
		utils.Log(utils.LevelError, "Encryption failed | Algorithm: %s | Type: %s | Target: %s | Latency: %v | Error: %v", algo, inputType, input, latency, err)
		return "", err
	}

	utils.Log(utils.LevelInfo, "Encryption successful | Algorithm: %s | Type: %s | Target: %s | Output: %q | Latency: %v", algo, inputType, input, output, latency)
	return output, nil
}

// DecryptMiddleware handles decryption with automatic file/string detection and logging.
func DecryptMiddleware(input string, key []byte, algo Algorithm) (string, error) {
	start := time.Now()
	inputType := "string"
	if info, err := os.Stat(input); err == nil && !info.IsDir() {
		inputType = "file"
	}

	utils.Log(utils.LevelInfo, "Starting decryption | Algorithm: %s | Type: %s | Target: %s", algo, inputType, input)

	var output string
	var err error

	if inputType == "file" {
		output, err = DecryptFilePath(input, key, algo)
	} else {
		output, err = DecryptString(input, key, algo)
	}

	latency := time.Since(start)

	if err != nil {
		utils.Log(utils.LevelError, "Decryption failed | Algorithm: %s | Type: %s | Target: %s | Latency: %v | Error: %v", algo, inputType, input, latency, err)
		return "", err
	}

	utils.Log(utils.LevelInfo, "Decryption successful | Algorithm: %s | Type: %s | Target: %s | Output: %q | Latency: %v", algo, inputType, input, output, latency)
	return output, nil
}

func EncryptString(plainText string, key []byte, algo Algorithm) (string, error) {
	var cipherText []byte
	var err error

	switch algo {
	case AES:
		cipherText, err = EncryptAES([]byte(plainText), key)
	case DES:
		cipherText, err = EncryptDES(key, []byte(plainText))
	case RC4:
		cipherText, err = EncryptRC4([]byte(plainText), key)
	default:
		utils.Log(utils.LevelWarning, "Attempted encryption with unknown algorithm: %q", algo)
		return "", fmt.Errorf("unknown algorithm: %q", algo)
	}
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func DecryptString(cipherTextB64 string, key []byte, algo Algorithm) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		utils.Log(utils.LevelWarning, "Failed to decode base64 input during decryption: %v", err)
		return "", fmt.Errorf("input is not valid base64 ciphertext: %w", err)
	}

	var plainText []byte
	switch algo {
	case AES:
		plainText, err = DecryptAES(cipherText, key)
	case DES:
		plainText, err = DecryptDES(key, cipherText)
	case RC4:
		plainText, err = DecryptRC4(cipherText, key)
	default:
		utils.Log(utils.LevelWarning, "Attempted decryption with unknown algorithm: %q", algo)
		return "", fmt.Errorf("unknown algorithm: %q", algo)
	}
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func EncryptFilePath(path string, key []byte, algo Algorithm) (string, error) {
	switch algo {
	case AES:
		return EncryptFileAES(path, key)
	case DES:
		return EncryptFileDES(key, path)
	case RC4:
		outPath := path + ".enc"
		if err := EncryptFileRC4(path, outPath, key); err != nil {
			return "", err
		}
		return outPath, nil
	default:
		utils.Log(utils.LevelWarning, "Attempted file encryption with unknown algorithm: %q", algo)
		return "", fmt.Errorf("unknown algorithm: %q", algo)
	}
}

func DecryptFilePath(path string, key []byte, algo Algorithm) (string, error) {
	switch algo {
	case AES:
		return DecryptFileAES(path, key)
	case DES:
		return DecryptFileDES(key, path)
	case RC4:
		outPath := path + ".dec"
		if err := DecryptFileRC4(path, outPath, key); err != nil { // or DecryptFileRC4 based on your implementation
			return "", err
		}
		return outPath, nil
	default:
		utils.Log(utils.LevelWarning, "Attempted file decryption with unknown algorithm: %q", algo)
		return "", fmt.Errorf("unknown algorithm: %q", algo)
	}
}