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