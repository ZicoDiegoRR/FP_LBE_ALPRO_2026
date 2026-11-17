package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Inisialisasi Gin router dengan default logger & recovery middleware
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "PMO Security API",
		})
	})

	// Jalankan server pada port 8080
	if err := r.Run(":8080"); err != nil {
		panic("Gagal menjalankan server: " + err.Error())
	}
}