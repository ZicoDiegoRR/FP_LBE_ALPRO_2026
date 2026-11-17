package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/security"
)

// AuthMiddleware memeriksa keberadaan dan validitas JWT pada header Authorization
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Akses ditolak: Header Authorization tidak ditemukan",
			})
			c.Abort() // Menghentikan eksekusi handler selanjutnya
			return
		}

		// Header harus berformat: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Format header Authorization salah (gunakan: Bearer <token>)",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validasi token kriptografis
		claims, err := security.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// Simpan data identitas pengguna ke context Gin agar bisa dibaca oleh handler di hilir
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next() // Lanjutkan ke handler utama
	}
}

// RequireRole adalah middleware Role-Based Access Control (RBAC)
func RequireRole(allowedRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != allowedRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Akses terlarang: Anda tidak memiliki hak akses yang cukup",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}