package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/config"
	"pmo-backend/internal/crypto"
	"pmo-backend/internal/database"
	"pmo-backend/internal/handler"
	"pmo-backend/internal/middleware"
	"pmo-backend/internal/repository"
)

func main() {
	// 1. Muat konfigurasi terpusat
	cfg := config.LoadConfig()

	// 2. Terapkan mode Gin
	gin.SetMode(cfg.GinMode)

	// 3. Hubungkan ke Database & Migrasi Skema
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Kesalahan database fatal: %v", err)
	}

	// 4. Inisialisasi Service & Repository Layer
	userRepo := repository.NewUserRepository(db)
	loanRepo := repository.NewLoanRepository(db)
	cryptoService := crypto.NewCryptoService()

	// 5. Inisialisasi Handler Layer
	authHandler := handler.NewAuthHandler(userRepo, cfg)
	loanHandler := handler.NewLoanHandler(loanRepo, cryptoService, cfg)

	// 6. Inisialisasi Gin Engine & Middleware Global
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// 7. Endpoint Health Check Publik
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"app":     cfg.AppName,
			"version": "1.0.0",
			"db":      cfg.DBDriver,
		})
	})

	// 8. Rute API v1
	apiV1 := r.Group("/api/v1")
	{
		// Rute Autentikasi Publik
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		// Rute Terproteksi (Wajib Membawa Token JWT Valid)
		protected := apiV1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Profil User
			protected.GET("/auth/me", authHandler.GetProfile)

			// Peminjaman & Kriptografi File
			loansGroup := protected.Group("/loans")
			{
				loansGroup.POST("", loanHandler.CreateLoan)
				loansGroup.GET("/my", loanHandler.GetUserLoans)
				loansGroup.GET("/:id/download", loanHandler.DownloadDecryptedFile)
			}

			// Rute Khusus Admin (Role-Based Access Control)
			adminGroup := protected.Group("/admin")
			adminGroup.Use(middleware.RequireRole("admin"))
			{
				adminGroup.GET("/loans", loanHandler.GetAllLoans)
			}
		}
	}

	// 9. Jalankan Server
	serverAddr := ":" + cfg.Port
	log.Printf("[INFO] Server berjalan di alamat http://localhost%s\n", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}