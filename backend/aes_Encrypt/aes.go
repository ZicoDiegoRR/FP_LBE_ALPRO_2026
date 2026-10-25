// AES 256
package aesmodule

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	//"fmt"
	"io"
)

func Encrypt(plainText, key []byte) ([]byte, error) {
	if len(key) !=32{
		return nil, errors.New("Key must be 32 Byte!")
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

// func GenerateRandomKey(size int) ([]byte, error) {
// 	key := make([]byte, size)
// 	_, err := rand.Read(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return key, nil
// }

func Decrypt(cipherText, key []byte) ([]byte, error) {
	if len(key) !=32{
		return nil, errors.New("Key must be 32 Byte!")
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

