package handlers

import (
	"os"
	"time"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"pmo-backend/db"
	"pmo-backend/encrypt"
	"pmo-backend/models"
)

// Maps a BorrowStatus id to the short type string the frontend already
// knows how to render an icon for (same categories as the dashboard).
func statusTypeForID(id string) string {
	switch id {
	case statusCancelled:
		return "cancelled"
	case statusPending:
		return "pending"
	case statusAccepted:
		return "accepted"
	case statusRejected:
		return "rejected"
	case statusOngoing:
		return "ongoing"
	case statusFinished:
		return "finished"
	case statusFailed:
		return "failed"
	default:
		return "none"
	}
}

type ProfileBorrowItem struct {
	ID                        string `json:"id"`
	Title                     string `json:"title"`
	BookedAt                  string `json:"bookedAt"`
	StatusMessage             string `json:"statusMessage"`
	StatusMessageUpdateTime   string `json:"statusMessageUpdateTime"`
	StatusType                string `json:"statusType"`
}

// GET /profile (requires auth)
func GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, APIResponse{false, "User not found"})
		return
	}

	var role models.Role
	db.DB.Where("id = ?", user.RoleID).First(&role) // role.Name stays "" if this fails, that's fine

	syncBorrowStatuses(userID)

	key := encrypt.GetEncryptionKey()
	algo := encrypt.GetAlgorithm()

	// Decrypt user profile fields
	decryptedName, err := encrypt.DecryptMiddleware(user.Name, key, algo)
	if err != nil {
		decryptedName = user.Name
	}
	decryptedEmail, err := encrypt.DecryptMiddleware(user.Email, key, algo)
	if err != nil {
		decryptedEmail = user.Email
	}

	// In GetProfile:
	var decryptedPfpURL string
	if user.PfpURL != "" {
    	decryptedPfpURL = "http://localhost:8080/pfp/serve/" + user.ID
	} else {
    	decryptedPfpURL = defaultPfpURL
	}

	var borrows []models.Borrow
	db.DB.Where("user_id = ?", userID).
		Order("statusmessageupdatetime DESC").
		Find(&borrows)

	items := make([]ProfileBorrowItem, 0, len(borrows))
	for _, b := range borrows {
		// Decrypt StatusMessage for each borrow item
		decStatusMsg, err := encrypt.DecryptMiddleware(b.StatusMessage, key, algo)
		if err != nil {
			decStatusMsg = b.StatusMessage
		}

		items = append(items, ProfileBorrowItem{
			ID:                        b.ID,
			Title:                     truncate(b.Title, activityTextLimit),
			BookedAt:                  b.BookedDatetime.Format("Jan 2, 2006 15:04"),
			StatusMessage:             truncate(decStatusMsg, activityTextLimit),
			StatusMessageUpdateTime:   b.StatusMessageUpdateTime.Format("Jan 2, 2006 15:04"),
			StatusType:                statusTypeForID(b.BorrowStatusID),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":           user.ID,
			"name":         decryptedName,
			"email":        decryptedEmail,
			"pfpUrl":       decryptedPfpURL,
			"roleName":     role.Name,
			"creationDate": user.CreationDate.Format("Jan 2, 2006"),
		},
		"borrowRequests": items,
	})
}

// POST /profile/pfp (requires auth)
func UploadProfilePicture(c *gin.Context) {
	userID := c.GetString("user_id")

	file, err := c.FormFile("pfp")
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{false, "No image file provided"})
		return
	}

	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to create uploads directory"})
		return
	}

	// 1. Save incoming file temporarily as raw
	rawFilename := userID + "_" + file.Filename
	rawPath := uploadsDir + "/raw_" + rawFilename
	if err := c.SaveUploadedFile(file, rawPath); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to save uploaded file"})
		return
	}

	key := encrypt.GetEncryptionKey()
	algo := encrypt.GetAlgorithm()

	// 2. Encrypt the file on disk using your encrypt middleware
	encryptedFilePath, err := encrypt.Middleware(rawPath, key, algo)
	os.Remove(rawPath) // Clean up the unencrypted raw file immediately
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to encrypt image file"})
		return
	}

	// 3. Encrypt the final file path string for database storage
	encryptedDbPath, err := encrypt.EncryptString(encryptedFilePath, key, algo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to encrypt file path for database"})
		return
	}

	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, APIResponse{false, "User not found"})
		return
	}

	user.PfpURL = encryptedDbPath
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to update profile picture in database"})
		return
	}

	// Public URL points to our dynamic decryption serving endpoint
	serveURL := "http://localhost:8080/pfp/serve/" + userID

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile picture encrypted and updated successfully",
		"pfpUrl":  serveURL,
	})
}

// GET /pfp/serve/:userId
func ServeProfilePicture(c *gin.Context) {
	requestedUserID := c.Param("userId")

	var user models.User
	if err := db.DB.Where("id = ?", requestedUserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.PfpURL == "" || user.PfpURL == defaultPfpURL {
		c.JSON(http.StatusNotFound, gin.H{"error": "No profile picture set"})
		return
	}

	key := encrypt.GetEncryptionKey()
	algo := encrypt.GetAlgorithm()

	// 1. Decrypt the database path string to get the encrypted file path on disk
	encryptedFilePath, err := encrypt.DecryptMiddleware(user.PfpURL, key, algo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt profile picture path"})
		return
	}

	// 2. Ensure ./temp directory exists
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp directory"})
		return
	}

	// 3. Define a unique target path inside ./temp using fmt.Sprintf to safely convert time to string
	tempTargetPath := fmt.Sprintf("%s/%s_%d.jpg", tempDir, requestedUserID, time.Now().UnixNano())

	// 4. Run DecryptMiddleware on the encrypted file path
	decryptedPath, err := encrypt.DecryptMiddleware(encryptedFilePath, key, algo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt image file"})
		return
	}

	// Read decrypted file bytes and copy into the ./temp folder so the cleanup scheduler tracks it
	inputData, err := os.ReadFile(decryptedPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read decrypted file"})
		return
	}

	if err := os.WriteFile(tempTargetPath, inputData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to temp folder"})
		return
	}

	// Clean up intermediate decryption artifact if your middleware left one behind
	if decryptedPath != tempTargetPath {
		_ = os.Remove(decryptedPath)
	}

	// Serve the file from ./temp so the background scheduler can purge it later
	c.File(tempTargetPath)
}

func StartTempCleanupScheduler(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			tempDir := "./temp"
			files, err := os.ReadDir(tempDir)
			if err != nil {
				continue
			}
			
			threshold := time.Now().Add(-5 * time.Minute) // Remove files older than 5 minutes
			for _, file := range files {
				info, err := file.Info()
				if err == nil && info.ModTime().Before(threshold) {
					_ = os.Remove(tempDir + "/" + file.Name())
				}
			}
		}
	}()
}