package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogLevel string

const (
	LevelInfo    LogLevel = "INFO"
	LevelWarning LogLevel = "WARN"
	LevelError   LogLevel = "ERROR"
)

var (
	logFile      *os.File
	fileLog      *log.Logger
	logMutex     sync.Mutex
	auditEnabled = true
	logPath      = getEnv("LOG_FILE_PATH", filepath.Join("logs", "crypto_audit.log"))
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// SetAuditEnabled mengaktifkan atau menonaktifkan penulisan audit log ke file (berguna saat benchmark).
func SetAuditEnabled(enabled bool) {
	logMutex.Lock()
	defer logMutex.Unlock()
	auditEnabled = enabled
}

// CloseLogFile menutup file log aktif secara aman.
func CloseLogFile() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		fileLog = nil
		return err
	}
	return nil
}

// SetLogFile mengubah path file log audit secara thread-safe.
func SetLogFile(path string) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
		fileLog = nil
	}

	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logFile = f
	logPath = path
	fileLog = log.New(logFile, "", 0)
	return nil
}

// Log menulis baris log terformat dengan timestamp dan level secara thread-safe.
func Log(level LogLevel, format string, v ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if !auditEnabled {
		return
	}

	if fileLog == nil {
		if dir := filepath.Dir(logPath); dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0755)
		}
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[CRITICAL] Gagal membuka file log %s: %v\n", logPath, err)
			return
		}
		logFile = f
		fileLog = log.New(logFile, "", 0)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, v...)
	fileLog.Printf("[%s] [%s] %s\n", timestamp, level, message)
}

// LogCryptoOperation mencatat audit log untuk setiap operasi enkripsi atau dekripsi.
func LogCryptoOperation(operation, algorithm string, sizeBytes int, duration time.Duration, err error) {
	if err != nil {
		Log(LevelError, "CRYPTO op=%s algo=%s bytes=%d latency=%s status=FAILED err=%v",
			operation, algorithm, sizeBytes, duration, err)
		return
	}
	Log(LevelInfo, "CRYPTO op=%s algo=%s bytes=%d latency=%s status=OK",
		operation, algorithm, sizeBytes, duration)
}
