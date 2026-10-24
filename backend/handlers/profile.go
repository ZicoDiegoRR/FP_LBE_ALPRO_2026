package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pmo-backend/db"
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
	ID                      string `json:"id"`
	Title                   string `json:"title"`
	BookedAt                string `json:"bookedAt"`
	StatusMessage           string `json:"statusMessage"`
	StatusMessageUpdateTime string `json:"statusMessageUpdateTime"`
	StatusType              string `json:"statusType"`
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

	var borrows []models.Borrow
	db.DB.Where("user_id = ?", userID).
		Order("statusmessageupdatetime DESC").
		Find(&borrows)

	items := make([]ProfileBorrowItem, 0, len(borrows))
	for _, b := range borrows {
		items = append(items, ProfileBorrowItem{
			ID:                      b.ID,
			Title:                   truncate(b.Title, activityTextLimit),
			BookedAt:                b.BookedDatetime.Format("Jan 2, 2006 15:04"),
			StatusMessage:           truncate(b.StatusMessage, activityTextLimit),
			StatusMessageUpdateTime: b.StatusMessageUpdateTime.Format("Jan 2, 2006 15:04"),
			StatusType:              statusTypeForID(b.BorrowStatusID),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":           user.ID,
			"name":         user.Name,
			"email":        user.Email,
			"pfpUrl":       user.PfpURL,
			"roleName":     role.Name,
			"creationDate": user.CreationDate.Format("Jan 2, 2006"),
		},
		"borrowRequests": items,
	})
}