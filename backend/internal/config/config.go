package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config menyimpan seluruh konfigurasi aplikasi
type Config struct {
	Port         string
	GinMode      string
	AppName      string
	DBDriver     string
	DBName       string
	JWTSecret    string
	CryptoKeyAES string
	CryptoKeyDES string
	CryptoKeyRC4 string
	StoragePath  string
}

// LoadConfig membaca environment variable dari file .env atau sistem OS
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("[INFO] File .env tidak ditemukan, menggunakan environment OS bawaan")
	}

	return &Config{
		Port:         getEnv("PORT", "8080"),
		GinMode:      getEnv("GIN_MODE", "debug"),
		AppName:      getEnv("APP_NAME", "PMO-Backend"),
		DBDriver:     getEnv("DB_DRIVER", "sqlite"),
		DBName:       getEnv("DB_NAME", "pmo.db"),
		JWTSecret:    getEnv("JWT_SECRET", "default_insecure_jwt_secret_change_me"),
		CryptoKeyAES: getEnv("CRYPTO_KEY_AES", "12345678901234567890123456789012"), // 32 byte
		CryptoKeyDES: getEnv("CRYPTO_KEY_DES", "12345678"),                         // 8 byte
		CryptoKeyRC4: getEnv("CRYPTO_KEY_RC4", "1234567890123456"),                 // 16 byte
		StoragePath:  getEnv("STORAGE_PATH", "./storage/encrypted"),
	}
}

// Helper untuk membaca env dengan nilai default jika kosong
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}