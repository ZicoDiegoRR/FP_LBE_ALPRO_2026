package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"pmo-backend/encrypt"
)

type BenchmarkTask struct {
	Name        string
	KeySize     int
	EncryptFunc func(data, key []byte) ([]byte, error)
	DecryptFunc func(data, key []byte) ([]byte, error)
}

type Payload struct {
	Label string
	Data  []byte
}

func main() {
	textSamples := []string{
		"pawsd",
		"Username",
		"data peminjaman",
	}

	// Isi dulu
	imagePath := "PATHMU"

	payloads := make([]Payload, 0, len(textSamples)+1)
	for i, sample := range textSamples {
		payloads = append(payloads, Payload{
			Label: fmt.Sprintf("Text %d", i+1),
			Data:  []byte(sample),
		})
	}

	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Printf("Gagal membaca gambar %q: %v\n", imagePath, err)
		return
	}
	payloads = append(payloads, Payload{
		Label: "Image",
		Data:  imageData,
	})

	tasks := []BenchmarkTask{
		{
			Name:        "AES-256-GCM",
			KeySize:     32,
			EncryptFunc: encrypt.EncryptAES,
			DecryptFunc: encrypt.DecryptAES,
		},
		{
			Name:    "DES-CBC",
			KeySize: 8,
			EncryptFunc: func(data, key []byte) ([]byte, error) {
				return encrypt.EncryptDES(key, data)
			},
			DecryptFunc: func(data, key []byte) ([]byte, error) {
				return encrypt.DecryptDES(key, data)
			},
		},
		{
			Name:        "RC4",
			KeySize:     16,
			EncryptFunc: encrypt.EncryptRC4,
			DecryptFunc: encrypt.DecryptRC4,
		},
	}

	iterations := 200

	for _, payload := range payloads {
		byteCount := len(payload.Data)
		sizeMB := float64(byteCount) / (1024 * 1024)

		fmt.Printf("--- Benchmarking Payload: %s (%d bytes) ---\n",
			payload.Label, byteCount)

		for _, task := range tasks {
			key := make([]byte, task.KeySize)
			if _, err := rand.Read(key); err != nil {
				fmt.Printf("Gagal membuat key %s: %v\n", task.Name, err)
				return
			}

			ciphertext, err := task.EncryptFunc(payload.Data, key)
			if err != nil {
				fmt.Printf("Enkripsi %s gagal: %v\n", task.Name, err)
				return
			}
			restored, err := task.DecryptFunc(ciphertext, key)
			if err != nil {
				fmt.Printf("Dekripsi %s gagal: %v\n", task.Name, err)
				return
			}
			if !bytes.Equal(restored, payload.Data) {
				fmt.Printf("Hasil dekripsi %s berbeda dari data asli\n", task.Name)
				return
			}

			encryptStart := time.Now()
			for i := 0; i < iterations; i++ {
				if _, err := task.EncryptFunc(payload.Data, key); err != nil {
					fmt.Printf("Benchmark enkripsi %s gagal: %v\n", task.Name, err)
					return
				}
			}
			encryptElapsed := time.Since(encryptStart)

			decryptStart := time.Now()
			for i := 0; i < iterations; i++ {
				if _, err := task.DecryptFunc(ciphertext, key); err != nil {
					fmt.Printf("Benchmark dekripsi %s gagal: %v\n", task.Name, err)
					return
				}
			}
			decryptElapsed := time.Since(decryptStart)

			encryptAvg := encryptElapsed.Seconds() / float64(iterations)
			decryptAvg := decryptElapsed.Seconds() / float64(iterations)

			fmt.Printf("%-15s | Encrypt: %8.2f MB/s, %.6f s | Decrypt: %8.2f MB/s, %.6f s\n",
				task.Name,
				sizeMB/encryptAvg, encryptAvg,
				sizeMB/decryptAvg, decryptAvg,
			)
		}
		fmt.Println()
	}
}