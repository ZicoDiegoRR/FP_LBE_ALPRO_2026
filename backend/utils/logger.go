package utils

import (
	"fmt"
	"log"
	"os"
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
	logFile  *os.File
	fileLog  *log.Logger
	logMutex sync.Mutex
	logPath  = GetEnv("LOG_FILE_PATH", "log-" + time.Now().Format("2006-01-02_15-04-05.000") + ".txt")
)

// SetLogFile sets a custom log file path.
func SetLogFile(path string) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		_ = logFile.Close()
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

// Log writes a formatted message with a specific log level.
func Log(level LogLevel, format string, v ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if fileLog == nil {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[CRITICAL] Failed to open log file: %v\n", err)
			return
		}
		logFile = f
		fileLog = log.New(logFile, "", 0)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)
	fileLog.Print(logLine)
}