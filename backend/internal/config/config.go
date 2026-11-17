package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config menyimpan seluruh konfigurasi aplikasi
type Config struct {
	Port    string
	GinMode string
	AppName string
}

// LoadConfig membaca environment variable dari file .env atau sistem OS
func LoadConfig() *Config {
	// Muat file .env jika ada (tidak fatal jika gagal di environment production)
	err := godotenv.Load()
	if err != nil {
		log.Println("[INFO] File .env tidak ditemukan, menggunakan environment OS bawaan")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Nilai default jika tidak disetel
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "PMO-Backend"
	}

	return &Config{
		Port:    port,
		GinMode: ginMode,
		AppName: appName,
	}
}