package main

import (
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
)

func main() {
	// Loads variables from .env into the environment. If the file doesn't
	// exist (e.g. in production where real env vars are set another way),
	// this just logs a notice and continues without failing.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on real environment variables")
	}

	db.Connect()

	// Periodically upgrades Accepted->Ongoing and Pending->Failed in the
	// background, independent of anyone actually loading a page. Interval
	// is overridable via STATUS_SYNC_INTERVAL_SECONDS; 30s by default —
	// borrow schedules only have minute precision, so this comfortably
	// keeps up without hammering the database.
	syncIntervalSeconds := 30
	if v := getEnv("STATUS_SYNC_INTERVAL_SECONDS", ""); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			syncIntervalSeconds = parsed
		}
	}
	go handlers.StartStatusSyncScheduler(time.Duration(syncIntervalSeconds) * time.Second)

	r := gin.Default()

	// Frontend (Next.js dev server) runs on a different port, so CORS must
	// explicitly allow it AND allow credentials for the session cookie to work.
	frontendOrigin := getEnv("FRONTEND_ORIGIN", "http://localhost:3000")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	// Cookie-based session store. SESSION_SECRET should be a long random
	// string in production; this fallback is fine for local dev only.
	secret := getEnv("SESSION_SECRET", "dev-only-secret-change-me")
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

	port := getEnv("GO_PORT", "8080")
	log.Println("Go backend running on http://localhost:" + port)
	r.Run(":" + port)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}