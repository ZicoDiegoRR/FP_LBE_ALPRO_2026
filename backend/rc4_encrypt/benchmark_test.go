package rc4_encrypt

import (
	"os"
	"path/filepath"
	"testing"
)

var benchmarkKey = []byte("benchmark-key-2026")

func loadBenchmarkFile(b *testing.B, name string) []byte {
	b.Helper()

	path := filepath.Join("..", "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("gagal membaca %s: %v", path, err)
	}

	if len(data) == 0 {
		b.Fatalf("file %s kosong", path)
	}

	return data
}

func BenchmarkRC4(b *testing.B) {
	files := []struct {
		name string
		file string
	}{
		{"Text", "sample.txt"},
		{"Image", "sample.jpg"},
		{"Video", "sample.mp4"},
	}

	for _, f := range files {
		b.Run(f.name+"/Encrypt", func(b *testing.B) {
			data := loadBenchmarkFile(b, f.file)
			b.SetBytes(int64(len(data)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := RC4Crypt(data, benchmarkKey)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(f.name+"/Decrypt", func(b *testing.B) {
			data := loadBenchmarkFile(b, f.file)

			encrypted, err := RC4Crypt(data, benchmarkKey)
			if err != nil {
				b.Fatal(err)
			}

			b.SetBytes(int64(len(data)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := RC4Crypt(encrypted, benchmarkKey)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
