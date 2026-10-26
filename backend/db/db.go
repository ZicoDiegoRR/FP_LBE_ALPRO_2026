package db

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pmo-backend/models"
)

var DB *gorm.DB

func Connect() {
	path := getEnv("DB_PATH", "pmo.db")
	// busy_timeout + WAL mengurangi error "database is locked"
	// karena ada scheduler yang menulis di latar belakang.
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"

	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open SQLite database: %v", err)
	}

	if err := database.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.BorrowStatus{},
		&models.Lab{},
		&models.Equipment{},
		&models.EquipmentStatus{},
		&models.LabEquipment{},
		&models.Borrow{},
		&models.EquipmentBorrowItem{},
		&models.StoredFile{}, // <-- Ditambahkan di sini secara bersih
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	seed(database)

	DB = database
	log.Println("Connected to SQLite:", path)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}