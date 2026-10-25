package rc4_encrypt

import (
	"crypto/rc4"
	"fmt"
	"os"
)

// RC4Crypt mengenkripsi atau mendekripsi data bytes menggunakan RC4.
// Fungsi ini ditujukan untuk demonstrasi dan pengujian akademik.
func RC4Crypt(data []byte, key []byte) ([]byte, error) {
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat RC4 cipher: %w", err)
	}

	result := make([]byte, len(data))
	cipher.XORKeyStream(result, data)

	return result, nil
}

// RC4EncryptFile membaca file, mengenkripsi isinya, lalu menyimpan hasilnya.
func RC4EncryptFile(inputPath, outputPath string, key []byte) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file: %w", err)
	}

	encrypted, err := RC4Crypt(data, key)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, encrypted, 0600); err != nil {
		return fmt.Errorf("gagal menyimpan file terenkripsi: %w", err)
	}

	return nil
}

// RC4DecryptFile membaca file terenkripsi dan menyimpan hasil dekripsinya.
func RC4DecryptFile(inputPath, outputPath string, key []byte) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file terenkripsi: %w", err)
	}

	decrypted, err := RC4Crypt(data, key)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, decrypted, 0600); err != nil {
		return fmt.Errorf("gagal menyimpan file hasil dekripsi: %w", err)
	}

	return nil
}
