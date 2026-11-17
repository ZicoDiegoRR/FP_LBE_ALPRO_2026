package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pmo-backend/internal/config"
	"pmo-backend/internal/model"
)

// ConnectDB membuka koneksi ke database dan melakukan migrasi skema tabel
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Menentukan driver database sesuai konfigurasi
	switch cfg.DBDriver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DBName), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	default:
		return nil, fmt.Errorf("driver database tidak didukung: %s", cfg.DBDriver)
	}

	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke database: %w", err)
	}

	log.Println("[INFO] Koneksi database berhasil diinisialisasi")

	// Auto-Migrasi skema database berdasarkan struct Model
	err = db.AutoMigrate(
		&model.User{},
		&model.Loan{},
	)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan auto-migrasi tabel: %w", err)
	}

	log.Println("[INFO] Auto-migrasi skema tabel (users, loans) selesai")

	return db, nil
}