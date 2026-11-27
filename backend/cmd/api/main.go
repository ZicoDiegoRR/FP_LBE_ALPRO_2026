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

	// 4. Inisialisasi Repository Layer
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrgRepository(db)
	loanRepo := repository.NewLoanRepository(db)
	vaultRepo := repository.NewVaultRepository(db)
	cryptoService := crypto.NewCryptoService()

	// 5. Inisialisasi Handler Layer
	authHandler := handler.NewAuthHandler(userRepo, cfg)
	orgHandler := handler.NewOrgHandler(orgRepo, cryptoService, cfg)
	loanHandler := handler.NewLoanHandler(loanRepo, orgRepo, cryptoService, cfg)
	vaultHandler := handler.NewVaultHandler(vaultRepo, cryptoService, cfg)

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
		// Autentikasi Publik
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		// Terproteksi (Wajib Login)
		protected := apiV1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/auth/me", authHandler.GetProfile)

			// --- MODUL ORGANISASI & KEANGGOTAAN ---
			protected.POST("/orgs/join", orgHandler.JoinOrganization)
			protected.GET("/orgs/my-membership", orgHandler.GetMyMembership)
			protected.GET("/equipments", orgHandler.GetEquipments)

			// Khusus Admin Organisasi
			adminOrgGroup := protected.Group("/admin/org")
			adminOrgGroup.Use(middleware.RequireRole("admin"))
			{
				adminOrgGroup.POST("", orgHandler.CreateOrganization)
				adminOrgGroup.GET("", orgHandler.GetMyOrganization)
				adminOrgGroup.PATCH("/schema", orgHandler.UpdateCustomSchema)
				adminOrgGroup.GET("/members", orgHandler.GetPendingMembers)
				adminOrgGroup.PATCH("/members/:id", orgHandler.VerifyMember)
				adminOrgGroup.POST("/equipments", orgHandler.AddEquipment)
				adminOrgGroup.GET("/loans", loanHandler.GetOrgLoans)
				adminOrgGroup.PATCH("/loans/:id/approve", loanHandler.ApproveLoan)
				adminOrgGroup.PATCH("/loans/:id/reject", loanHandler.RejectLoan)
				adminOrgGroup.POST("/loans/:id/return-confirm", loanHandler.ConfirmReturn)
			}

			// --- MODUL PEMINJAMAN ALAT (USER & ADMIN) ---
			loansGroup := protected.Group("/loans")
			{
				loansGroup.POST("", loanHandler.CreateLoan)
				loansGroup.GET("/my", loanHandler.GetUserLoans)
				loansGroup.POST("/:id/return-request", loanHandler.RequestReturn)
				loansGroup.GET("/:id/download", loanHandler.DownloadDecryptedFile)
			}

			// --- MODUL BRANKAS KRIPTOGRAFI, DIGITAL SIGNATURE, & FILE SHARING ---
			vaultGroup := protected.Group("/vault")
			{
				vaultGroup.POST("/keys/generate", vaultHandler.GenerateKeyPair)
				vaultGroup.GET("/files", vaultHandler.GetMyVaultFiles)
				vaultGroup.POST("/files", vaultHandler.UploadVaultFile)
				vaultGroup.POST("/files/:id/sign", vaultHandler.SignDocument)
				vaultGroup.GET("/files/:id/verify", vaultHandler.VerifySignature)
				
				// Secure Data Sharing
				vaultGroup.POST("/share/request", vaultHandler.RequestAccess)
				vaultGroup.POST("/share/requests/:id/approve", vaultHandler.ApproveShareRequest)
				vaultGroup.POST("/share/requests/:id/revoke", vaultHandler.RevokeShareAccess)
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