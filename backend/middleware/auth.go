package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// RequireAuth blocks the request unless a valid session cookie is present,
// and makes the logged-in user's id available via c.Get("user_id").
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not logged in"})
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Set("user_name", session.Get("user_name"))
		c.Next()
	}
}
