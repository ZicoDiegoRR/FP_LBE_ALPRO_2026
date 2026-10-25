package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pmo-backend/db"
	"pmo-backend/encrypt"
	"pmo-backend/models"
	"pmo-backend/utils"
)

type SubmitBorrowRequest struct {
	Title        string   `json:"title" binding:"required"`
	LabID        string   `json:"labId" binding:"required"`
	Reason       string   `json:"reason" binding:"required"`
	ScheduledAt  string   `json:"scheduledAt" binding:"required"` // "2006-01-02T15:04" (datetime-local)
	Days         int      `json:"days"`
	Hours        int      `json:"hours"`
	Minutes      int      `json:"minutes"`
	Seconds      int      `json:"seconds"`
	EquipmentIDs []string `json:"equipmentIds" binding:"required,min=1"`
}

var scheduleLocation = time.FixedZone("WIB", 7*60*60)

func parseScheduledAt(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02T15:04", value, scheduleLocation)
}

func asScheduleLocation(t time.Time) time.Time {
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
		scheduleLocation,
	)
}

// POST /borrow (requires auth)
func SubmitBorrow(c *gin.Context) {
	userID := c.GetString("user_id")

	var req SubmitBorrowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{false, "Missing or invalid fields"})
		return
	}

	scheduled, err := parseScheduledAt(req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{false, "Invalid scheduled date/time"})
		return
	}

	totalSeconds := req.Days*86400 + req.Hours*3600 + req.Minutes*60 + req.Seconds
	if totalSeconds <= 0 {
		c.JSON(http.StatusBadRequest, APIResponse{false, "Duration must be more than 0"})
		return
	}

	// Fetch key and algorithm directly from the encrypt package
	key := encrypt.GetEncryptionKey()
	algo := encrypt.GetAlgorithm()

	// Encrypt the reason provided in the request
	encryptedReason, err := encrypt.Middleware(req.Reason, key, algo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to encrypt reason"})
		return
	}

	// Encrypt the initial status message
	rawStatusMessage := "Request is pending for approval."
	encryptedStatusMessage, err := encrypt.Middleware(rawStatusMessage, key, algo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to encrypt status message"})
		return
	}

	now := time.Now()
	borrow := models.Borrow{
		ID:                      utils.GenerateID(),
		ScheduledDatetime:       scheduled,
		BookedDatetime:          now,
		Reason:                  encryptedReason,
		BorrowDurationSeconds:   totalSeconds,
		BorrowStatusID:          statusPending,
		UserID:                  userID,
		FinishedAt:              nil,
		Title:                   req.Title,
		StatusMessage:           encryptedStatusMessage,
		StatusMessageUpdateTime: now,
	}

	if err := db.DB.Create(&borrow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not create borrow request"})
		return
	}

	items := make([]models.EquipmentBorrowItem, 0, len(req.EquipmentIDs))
	for _, eqID := range req.EquipmentIDs {
		items = append(items, models.EquipmentBorrowItem{
			BorrowID:       borrow.ID,
			LabEquipmentID: eqID,
		})
	}
	if err := db.DB.Create(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Borrow created, but linking equipment failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Borrow request submitted", "borrowId": borrow.ID})
}