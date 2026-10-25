package main

import (
	"crypto/rand"
	//"encoding/json"
	"fmt"
	//"os"
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

func main() {
	payloadSizes := map[string]int64{
		"1 KB":  1024,
		"64 KB": 64 * 1024,
		"1 MB":  1024 * 1024,
		//"100 MB": 1024 * 1024,
	}

	tasks := []BenchmarkTask{
		//{Name: "DES-{block}", KeySize: 8, EncryptFunc: cryptoalgos.Encrypt},
		{Name: "AES-256-GCM", KeySize: 32, EncryptFunc: encrypt.EncryptAES},
		//{Name: "RC4", KeySize: 16, EncryptFunc: cryptoalgos.Encrypt},
	}

	iterations := 200

	for sizeLabel, byteCount := range payloadSizes {
		payload := make([]byte, byteCount)
		rand.Read(payload)

		fmt.Printf("--- Benchmarking Payload: %s ---\n", sizeLabel)

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