package aesmodule

import (
    "bytes"
    "crypto/rand"
    "testing"
)

func TestEncryptDecrypt(t *testing.T) {
    key := make([]byte, 32)
    if _, err := rand.Read(key); err != nil {
        t.Fatal(err)
    }

    original := []byte("Hello from AES")

    encrypted, err := Encrypt(original, key)
    if err != nil {
        t.Fatal(err)
    }

    decrypted, err := Decrypt(encrypted, key)
    if err != nil {
        t.Fatal(err)
    }

    if !bytes.Equal(decrypted, original) {
        t.Fatalf("hasil dekripsi berbeda: %q", decrypted)
    }

    t.Logf("Encrypted (hex): %x", encrypted)
    t.Logf("Decrypted: %s", decrypted)
}

func TestRejectInvalidKey(t *testing.T) {
    key := make([]byte, 31)

    if _, err := Encrypt([]byte("Hello from AES"), key); err == nil {
        t.Fatal("Encrypt menerima key 31 byte")
    }

    if _, err := Decrypt([]byte("data"), key); err == nil {
        t.Fatal("Decrypt menerima key 31 byte")
    }
}