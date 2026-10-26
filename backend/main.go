package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"pmo-backend/db"
	"pmo-backend/handlers"
	"pmo-backend/middleware"
	"pmo-backend/utils"
)

func main() {
	// 1. Define command-line flags (NOTICE: no leading '--' in flag names)
	safeTimestamp := time.Now().Format("2006-01-02_15-04-05.000") // Safe for file systems
	algoType := flag.String("encrypt-algo", "aes", "Encryption algorithm type (aes, des, rc4).")
	logPath := flag.String("log-path", "log-"+safeTimestamp+".txt", "Log file path.")
	keyEncrypt := flag.String("encrypt-key", "", "Encryption key.")
	
	// Parse the flags
	flag.Parse()

	// 2. Loads variables from .env into the environment.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on real environment variables")
	}

	// 3. Pass flag values to your modules / environment
	if *logPath != "" {
		_ = os.Setenv("LOG_FILE_PATH", *algoType)
	}

	if *keyEncrypt != "" {
		_ = os.Setenv("ENCRYPTION_SECRET_KEY", *keyEncrypt)
	}
	if *algoType != "" {
		_ = os.Setenv("ENCRYPTION_ALGO", *algoType)
	}

	db.Connect()

	// Periodically upgrades Accepted->Ongoing and Pending->Failed in the background
	syncIntervalSeconds := 30
	if v := utils.GetEnv("STATUS_SYNC_INTERVAL_SECONDS", ""); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			syncIntervalSeconds = parsed
		}
	}
	go handlers.StartStatusSyncScheduler(time.Duration(syncIntervalSeconds) * time.Second)

	r := gin.Default()

	frontendOrigin := utils.GetEnv("FRONTEND_ORIGIN", "http://localhost:3000")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	secret := utils.GetEnv("SESSION_SECRET", "dev-only-secret-change-me")
	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24, // 1 day
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("pmo_session", store))

	r.POST("/login", handlers.Login)
	r.POST("/signup", handlers.Signup)
	r.GET("/me", handlers.Me)
	r.POST("/logout", handlers.Logout)
	r.GET("/dashboard", middleware.RequireAuth(), handlers.GetDashboard)
	r.GET("/labs", middleware.RequireAuth(), handlers.ListLabs)
	r.GET("/labs/:id/equipment", middleware.RequireAuth(), handlers.GetLabEquipment)
	r.POST("/borrow", middleware.RequireAuth(), handlers.SubmitBorrow)
	r.GET("/profile", middleware.RequireAuth(), handlers.GetProfile)
	r.GET("/borrow/:id", middleware.RequireAuth(), handlers.GetBorrowDetail)
	r.POST("/borrow/:id/cancel", middleware.RequireAuth(), handlers.CancelBorrow)
	r.Static("/uploads", "./uploads") // Serves uploaded files publicly
	r.POST("/profile/pfp", middleware.RequireAuth(), handlers.UploadProfilePicture)
	r.GET("/pfp/serve/:userId", handlers.ServeProfilePicture)

	// === TAMBAHAN RUTE FILE ENKRIPSI (LANGKAH 4) ===
	r.POST("/files", middleware.RequireAuth(), handlers.UploadFile)
	r.GET("/files", middleware.RequireAuth(), handlers.ListFiles)
	r.GET("/files/:id", middleware.RequireAuth(), handlers.DownloadFile)
	// ===============================================

	port := utils.GetEnv("GO_PORT", "8080")
	log.Println("Go backend running on http://localhost:" + port)
	go handlers.StartTempCleanupScheduler(5 * time.Minute)
	r.Run(":" + port)
}