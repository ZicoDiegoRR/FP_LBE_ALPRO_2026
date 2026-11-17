package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config menyimpan seluruh konfigurasi aplikasi
type Config struct {
	Port     string
	GinMode  string
	AppName  string
	DBDriver string
	DBName   string
}

// LoadConfig membaca environment variable dari file .env atau sistem OS
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("[INFO] File .env tidak ditemukan, menggunakan environment OS bawaan")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "PMO-Backend"
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "pmo.db"
	}

	return &Config{
		Port:     port,
		GinMode:  ginMode,
		AppName:  appName,
		DBDriver: dbDriver,
		DBName:   dbName,
	}
}