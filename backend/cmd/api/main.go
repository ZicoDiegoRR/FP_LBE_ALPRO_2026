package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pmo-backend/internal/config"
)

func main() {
	// 1. Muat konfigurasi terpusat
	cfg := config.LoadConfig()

	// 2. Terapkan mode Gin sesuai konfigurasi
	gin.SetMode(cfg.GinMode)

	// 3. Inisialisasi router
	r := gin.Default()

	// 4. Endpoint Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"app":     cfg.AppName,
			"version": "1.0.0",
		})
	})

	// 5. Jalankan server dengan port dari config
	serverAddr := ":" + cfg.Port
	if err := r.Run(serverAddr); err != nil {
		panic("Gagal menjalankan server: " + err.Error())
	}
}