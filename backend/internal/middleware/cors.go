package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware mengatur header Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mengizinkan origin frontend (Next.js default port 3000)
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Pada production ketat bisa diganti http://localhost:3000
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		// Jika request adalah pre-flight (OPTIONS), langsung kembalikan status OK
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}