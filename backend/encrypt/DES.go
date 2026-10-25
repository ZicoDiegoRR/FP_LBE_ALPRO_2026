package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	// "encoding/hex"
	"errors"
	"fmt"

)

// Buat padding apabila melebihi 7
func pad(b []byte) []byte {
	n := des.BlockSize - len(b)%des.BlockSize
	return append(b, bytes.Repeat([]byte{byte(n)}, n)...)
}

func unpad(b []byte) ([]byte, error) {
	if len(b) == 0 || len(b)%des.BlockSize != 0 {
		return nil, errors.New("bad length")
	}
	n := int(b[len(b)-1])
	if n == 0 || n > des.BlockSize || !bytes.Equal(b[len(b)-n:], bytes.Repeat([]byte{byte(n)}, n)) {
		return nil, errors.New("bad padding")
	}
	return b[:len(b)-n], nil
}

func encryptCBC(key, plaintext []byte) ([]byte, error) {
	block, err := des.NewCipher(key) // key must be exactly 8 bytes
	if err != nil {
		return nil, err
	}
	p := pad(plaintext)
	out := make([]byte, des.BlockSize+len(p))
	iv := out[:des.BlockSize]
	if _, err := rand.Read(iv); err != nil { // fresh random IV per message
		return nil, err
	}
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out[des.BlockSize:], p)
	return out, nil // IV || ciphertext
}

func decryptCBC(key, data []byte) ([]byte, error) {
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
	return unpad(out)
}




func main() {
	// Known-answer test: one raw DES block
	key, _ := hex.DecodeString("133457799BBCDFF1")
	pt, _ := hex.DecodeString("0123456789ABCDEF")
	block, _ := des.NewCipher(key)
	ct := make([]byte, 8)
	block.Encrypt(ct, pt)
	fmt.Println("block:", hex.EncodeToString(ct))

	// Parity bits are ignored: flip the low bit of every key byte
	key2 := make([]byte, 8)
	for i := range key {
		key2[i] = key[i] ^ 1
	}
	block2, _ := des.NewCipher(key2)
	ct2 := make([]byte, 8)
	block2.Encrypt(ct2, pt)
	fmt.Println("same with flipped parity:", bytes.Equal(ct, ct2))

	// CBC round trip
	msg := []byte("legacy payload")
	enc, _ := encryptCBC(key, msg)
	fmt.Println("cbc bytes:", len(enc))
	dec, err := decryptCBC(key, enc)
	fmt.Println("decrypted:", string(dec), err)

	_, err = des.NewCipher([]byte("short"))
	fmt.Println(err)
}

