package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAuditLoggerThreadSafe(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "audit_test.log")
	if err := SetLogFile(logFile); err != nil {
		t.Fatalf("SetLogFile gagal: %v", err)
	}
	defer CloseLogFile()
	SetAuditEnabled(true)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			LogCryptoOperation("ENCRYPT", "AES", 128+idx, 2*time.Millisecond, nil)
		}(i)
	}
	LogCryptoOperation("DECRYPT", "DES", 64, time.Millisecond, errors.New("tampered"))
	wg.Wait()

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("gagal membaca file log: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "CRYPTO op=ENCRYPT algo=AES") {
		t.Fatalf("entri log ENCRYPT AES tidak ditemukan:\n%s", content)
	}
	if !strings.Contains(content, "CRYPTO op=DECRYPT algo=DES") || !strings.Contains(content, "status=FAILED") {
		t.Fatalf("entri log DECRYPT FAILED tidak ditemukan:\n%s", content)
	}
}
