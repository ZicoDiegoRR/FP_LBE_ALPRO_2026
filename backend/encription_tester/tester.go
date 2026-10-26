package main

import (
	"crypto/rand"
	//"encoding/json"
	"fmt"
	"os"
	"time"

	// Ini diganti dirnya
	"pmo-backend/encrypt"
	//"backend/.."
)

type BenchmarkTask struct {
	Name        string
	KeySize     int
	EncryptFunc func(data, key []byte) ([]byte, error)
}

type Payload struct {
	Label string
	Data  []byte
}
func main() {
	text_path := []string{
		"pawsd",
		"Username",
		"data peminjaman",
	}
	imagePath:="path"

	payloads := make([]Payload, 0, len(textSamples)+1)

	for i, sample := range textSamples {
		payloads = append(payloads, Payload{
			Label: fmt.Sprintf("Text %d", i+1),
			Data:  []byte(sample),
		})
	}
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Printf("Path salah %q: %v\n", imagePath, err)
		return
	}
	payloads = append(payloads, Payload{
		Label: "Image",
		Data:  imageData,
	})
	tasks := []BenchmarkTask{
		//{Name: "DES-{block}", KeySize: 8, EncryptFunc: cryptoalgos.Encrypt},
		{Name: "AES-256-GCM", KeySize: 32, EncryptFunc: encrypt.EncryptAES},
		//{Name: "RC4", KeySize: 16, EncryptFunc: cryptoalgos.Encrypt},
	}

	iterations := 200

	for _, payload := range payloads {
		byteCount := int64(len(payload.Data))
		rand.Read(payload)

		fmt.Printf("--- Benchmarking Payload:%s (%d bytes) ---\n",
			payload.Label, byteCount )

		for _, task := range tasks {
			key := make([]byte, task.KeySize)
			rand.Read(key)

			// Warmup run
			_, _ = task.EncryptFunc(payload, key)

			start := time.Now()
			for i := 0; i < iterations; i++ {
				_, err := task.EncryptFunc(payload, key)
				if err != nil {
					fmt.Printf("Error running %s: %v\n", task.Name, err)
					break
				}
			}
			elapsed := time.Since(start)

			avgSec := elapsed.Seconds() / float64(iterations)
			sizeMB := float64(byteCount) / (1024 * 1024)
			throughputMBps := sizeMB / avgSec

			fmt.Printf("%-15s | Throughput: %8.2f MB/s | Latency: %.6f s\n",
				task.Name, throughputMBps, avgSec)
		}
		fmt.Println()
	}
}