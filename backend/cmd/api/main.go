package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/config"
	"pmo-backend/internal/database"
)

func main() {
	// 1. Muat konfigurasi terpusat
	cfg := config.LoadConfig()

	// 2. Terapkan mode Gin
	gin.SetMode(cfg.GinMode)

	// 3. Hubungkan ke Database & Jalankan Migrasi
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Kesalahan database fatal: %v", err)
	}
	_ = db // Disimpan untuk di-passing ke repository/handler nanti

	// 4. Inisialisasi router Gin
	r := gin.Default()

	// 5. Endpoint Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"app":     cfg.AppName,
			"version": "1.0.0",
			"db":      cfg.DBDriver,
		})
	})

	// 6. Jalankan server
	serverAddr := ":" + cfg.Port
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}