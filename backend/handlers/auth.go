package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"pmo-backend/db"
	"pmo-backend/encrypt"
	"pmo-backend/models"
	"pmo-backend/utils"
)

// Every new sign-up gets this default profile picture unless they upload their own later.
const defaultPfpURL = "https://upload.wikimedia.org/wikipedia/commons/a/ac/Default_pfp.jpg?utm_source=commons.wikimedia.org&utm_campaign=index&utm_content=original"

// Every new sign-up gets this role unless DEFAULT_ROLE_ID overrides it.
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
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{false, "Username/email and password are required"})
		return
	}

	key := encrypt.GetEncryptionKey()
	algo := encrypt.GetAlgorithm()

	var matchedUser models.User
	var decryptedUsername string
	var decryptedEmail string

	if algo == encrypt.RC4 {
		// RC4 is deterministic: encrypting the input directly yields the stored ciphertext
		encryptedInput, err := encrypt.Middleware(req.UsernameOrEmail, key, algo)
		if err != nil {
			c.JSON(http.StatusUnauthorized, APIResponse{false, "Invalid credentials"})
			return
		}

		result := db.DB.Where("name = ? OR email = ?", encryptedInput, encryptedInput).First(&matchedUser)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, APIResponse{false, "Invalid credentials"})
			return
		}

		// Decrypt user fields for session and response
		decryptedUsername, _ = encrypt.DecryptMiddleware(matchedUser.Name, key, algo)
		decryptedEmail, _ = encrypt.DecryptMiddleware(matchedUser.Email, key, algo)

	} else {
		// AES and DES use randomized IVs, requiring iteration and decryption
		var users []models.User
		if err := db.DB.Find(&users).Error; err != nil {
			c.JSON(http.StatusUnauthorized, APIResponse{false, "Invalid credentials"})
			return
		}

		var foundUser *models.User
		for _, user := range users {
			decName, errName := encrypt.DecryptMiddleware(user.Name, key, algo)
			decEmail, errEmail := encrypt.DecryptMiddleware(user.Email, key, algo)

			if (errName == nil && decName == req.UsernameOrEmail) || (errEmail == nil && decEmail == req.UsernameOrEmail) {
				foundUser = &user
				decryptedUsername = decName
				decryptedEmail = decEmail
				break
			}
		}

		if foundUser == nil {
			c.JSON(http.StatusUnauthorized, APIResponse{false, "Invalid credentials"})
			return
		}
		matchedUser = *foundUser
	}

	// Verify password hash
	if err := bcrypt.CompareHashAndPassword([]byte(matchedUser.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, APIResponse{false, "Invalid credentials"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", matchedUser.ID)
	session.Set("user_name", decryptedUsername)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not start session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"user":    gin.H{"id": matchedUser.ID, "name": decryptedUsername, "email": decryptedEmail},
	})
}

// POST /signup
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

	key := encrypt.GetEncryptionKey()
	algo := encrypt.GetAlgorithm()

	// Encrypt username and email before saving
	encryptedUsername, err := encrypt.Middleware(req.Username, key, algo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to encrypt username"})
		return
	}

	encryptedEmail, err := encrypt.Middleware(req.Email, key, algo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to encrypt email"})
		return
	}

	// Generate bcrypt password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to process password"})
		return
	}

	user := models.User{
		ID:           utils.GenerateID(),
		Name:         encryptedUsername,
		CreationDate: time.Now(),
		Email:        encryptedEmail,
		Password:     string(hashedPassword),
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
func Me(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, APIResponse{false, "Not logged in"})
		return
	}

	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, APIResponse{false, "User not found"})
		return
	}

	key := encrypt.GetEncryptionKey()
	algo := encrypt.GetAlgorithm()

	decryptedName, _ := encrypt.DecryptMiddleware(user.Name, key, algo)
	decryptedEmail, _ := encrypt.DecryptMiddleware(user.Email, key, algo)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    gin.H{"id": user.ID, "name": decryptedName, "email": decryptedEmail},
	})
}

// POST /logout
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, APIResponse{true, "Logged out"})
}