package rc4_encrypt

import (
	"fmt"
	"testing"
)

func BenchmarkRC4(b *testing.B) {
	key := []byte("benchmark-key")

	testCases := []struct {
		name string
		size int
	}{
		{"Small_1KiB", 1 * 1024},
		{"Medium_1MiB", 1 * 1024 * 1024},
		{"Large_10MiB", 10 * 1024 * 1024},
	}

	for _, tc := range testCases {
		b.Run(tc.name+"/Encrypt", func(b *testing.B) {
			benchmarkRC4Crypt(b, key, tc.size)
		})

		b.Run(tc.name+"/Decrypt", func(b *testing.B) {
			plaintext := make([]byte, tc.size)

			// Siapkan ciphertext sebelum timer dimulai.
			ciphertext, err := RC4Crypt(plaintext, key)
			if err != nil {
				b.Fatal(err)
			}

			b.SetBytes(int64(tc.size))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := RC4Crypt(ciphertext, key)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchmarkRC4Crypt(b *testing.B, key []byte, size int) {
	plaintext := make([]byte, size)

	// Hindari input yang seluruhnya nol.
	for i := range plaintext {
		plaintext[i] = byte(i % 251)
	}

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := RC4Crypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

var _ = fmt.Sprintf
