package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"pmo-backend/db"
	"pmo-backend/models"
	"pmo-backend/utils"
)

// Every new sign-up gets this default profile picture unless they upload
// their own later (no upload flow exists yet).
const defaultPfpURL = "https://upload.wikimedia.org/wikipedia/commons/a/ac/Default_pfp.jpg?utm_source=commons.wikimedia.org&utm_campaign=index&utm_content=original"

// Every new sign-up gets this role unless DEFAULT_ROLE_ID overrides it.
// This is your existing "Member" role's id.
const fallbackDefaultRoleID = "F4L9SOWPVI7465NEOBM7"

func defaultRoleID() string {
	if v := os.Getenv("DEFAULT_ROLE_ID"); v != "" {
		return v
	}
	return fallbackDefaultRoleID
}

type LoginRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type SignupRequest struct {
	Username        string `json:"username" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// POST /login
// On success, stores the user's id/name in a signed session cookie.
// The frontend never needs to touch localStorage for this.
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{false, "Username/email and password are required"})
		return
	}

	var user models.User
	result := db.DB.Where(`name = ? OR email = ?`, req.UsernameOrEmail, req.UsernameOrEmail).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, APIResponse{false, "Invalid credentials"})
		return
	}

	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, APIResponse{false, "Invalid credentials"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("user_name", user.Name)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not start session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"user":    gin.H{"id": user.ID, "name": user.Name, "email": user.Email},
	})
}

// POST /signup
// Inserts a new row into "User" with a generated id and the default
// Member role. Password is stored as plain text (demo purposes only).
func Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{false, "All fields are required"})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, APIResponse{false, "Passwords do not match"})
		return
	}

	user := models.User{
		ID:           utils.GenerateID(),
		Name:         req.Username,
		CreationDate: time.Now(),
		Email:        req.Email,
		Password:     req.Password,
		PfpURL:       defaultPfpURL,
		RoleID:       defaultRoleID(),
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, APIResponse{false, "Username or email already exists"})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{true, "Account created successfully"})
}

// GET /me
// Lets the frontend check "is anyone logged in" by reading the session
// cookie, instead of keeping a copy of the user in localStorage.
func Me(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, APIResponse{false, "Not logged in"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    gin.H{"id": userID, "name": session.Get("user_name")},
	})
}

// POST /logout
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, APIResponse{true, "Logged out"})
}
