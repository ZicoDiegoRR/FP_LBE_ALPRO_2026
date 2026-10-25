package encrypt

import (
	"crypto/rc4"
	"fmt"
	"os"
)

// EncryptRC4 mengenkripsi atau mendekripsi data bytes menggunakan RC4
// (XOR stream cipher: the same function does both directions).
// Fungsi ini ditujukan untuk demonstrasi dan pengujian akademik.
func EncryptRC4(data []byte, key []byte) ([]byte, error) {
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat RC4 cipher: %w", err)
	}

	result := make([]byte, len(data))
	cipher.XORKeyStream(result, data)

	return result, nil
}

// DecryptRC4 is identical to EncryptRC4 — RC4 is a symmetric XOR stream
// cipher, so encryption and decryption are the same operation. Kept as a
// separate name for readability at call sites.
func DecryptRC4(data []byte, key []byte) ([]byte, error) {
	return EncryptRC4(data, key)
}

// EncryptFileRC4 membaca file, mengenkripsi isinya, lalu menyimpan hasilnya.
func EncryptFileRC4(inputPath, outputPath string, key []byte) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file: %w", err)
	}

	encrypted, err := EncryptRC4(data, key)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, encrypted, 0600); err != nil {
		return fmt.Errorf("gagal menyimpan file terenkripsi: %w", err)
	}

	return nil
}

// DecryptFileRC4 membaca file terenkripsi dan menyimpan hasil dekripsinya.
func DecryptFileRC4(inputPath, outputPath string, key []byte) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file terenkripsi: %w", err)
	}

	decrypted, err := DecryptRC4(data, key)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, decrypted, 0600); err != nil {
		return fmt.Errorf("gagal menyimpan file hasil dekripsi: %w", err)
	}

	return nil
}