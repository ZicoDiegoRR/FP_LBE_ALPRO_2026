package crypto

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"pmo-backend/internal/utils"
)

type benchAlgo struct {
	name string
	key  []byte
}

var benchmarkAlgorithms = []benchAlgo{
	{name: "AES", key: []byte("kunci-rahasia-pmo-harus-32-byte!")},
	{name: "DES", key: []byte("12345678")},
	{name: "RC4", key: []byte("1234567890123456")},
}

type benchPayload struct {
	name string
	data []byte
}

func loadBenchmarkPayloads(b *testing.B) []benchPayload {
	b.Helper()

	samplePath := filepath.Join("..", "..", "testdata", "sample.txt")
	sampleBytes, err := os.ReadFile(samplePath)
	if err != nil {
		b.Fatalf("gagal membaca sample testdata (%s): %v", samplePath, err)
	}

	makeRandom := func(size int) []byte {
		buf := make([]byte, size)
		if _, err := rand.Read(buf); err != nil {
			b.Fatalf("gagal membuat data acak %d byte: %v", size, err)
		}
		return buf
	}

	return []benchPayload{
		{name: "sample_txt", data: sampleBytes},
		{name: "1KB", data: makeRandom(1024)},
		{name: "64KB", data: makeRandom(64 * 1024)},
		{name: "1MB", data: makeRandom(1024 * 1024)},
	}
}

func BenchmarkEncrypt(b *testing.B) {
	utils.SetAuditEnabled(false)
	defer utils.SetAuditEnabled(true)

	svc := NewCryptoService()
	payloads := loadBenchmarkPayloads(b)

	for _, p := range payloads {
		for _, algo := range benchmarkAlgorithms {
			b.Run(algo.name+"/"+p.name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(p.data)))

				// Ukur overhead ukuran ciphertext satu kali sebelum timer utama
				sampleEnc, err := svc.Encrypt(p.data, algo.name, algo.key)
				if err != nil {
					b.Fatalf("Encrypt %s gagal: %v", algo.name, err)
				}
				b.ReportMetric(float64(len(sampleEnc)-len(p.data)), "overhead_B")

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					out, err := svc.Encrypt(p.data, algo.name, algo.key)
					if err != nil {
						b.Fatalf("Encrypt %s gagal: %v", algo.name, err)
					}
					if len(out) == 0 {
						b.Fatal("ciphertext kosong")
					}
				}
			})
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	utils.SetAuditEnabled(false)
	defer utils.SetAuditEnabled(true)

	svc := NewCryptoService()
	payloads := loadBenchmarkPayloads(b)

	for _, p := range payloads {
		for _, algo := range benchmarkAlgorithms {
			b.Run(algo.name+"/"+p.name, func(b *testing.B) {
				encrypted, err := svc.Encrypt(p.data, algo.name, algo.key)
				if err != nil {
					b.Fatalf("setup Encrypt %s gagal: %v", algo.name, err)
				}

				// Validasi round-trip satu kali sebelum loop benchmark
				checkPlain, err := svc.Decrypt(encrypted, algo.name, algo.key)
				if err != nil {
					b.Fatalf("verifikasi Decrypt %s gagal: %v", algo.name, err)
				}
				if !bytes.Equal(checkPlain, p.data) {
					b.Fatalf("hasil dekripsi %s tidak identik dengan plaintext", algo.name)
				}

				b.ReportAllocs()
				b.SetBytes(int64(len(p.data)))
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					plain, err := svc.Decrypt(encrypted, algo.name, algo.key)
					if err != nil {
						b.Fatalf("Decrypt %s gagal: %v", algo.name, err)
					}
					if len(plain) != len(p.data) {
						b.Fatalf("panjang plaintext tidak cocok: dapat %d, harap %d", len(plain), len(p.data))
					}
				}
			})
		}
	}
}
