package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pmo-backend/db"
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

// datetime-local inputs send a plain wall-clock string with no timezone
// info (e.g. "2026-09-20T14:00") — time.Parse would default that to UTC,
// which is wrong if the person entering it is anywhere else. Assuming a
// single-timezone deployment (WIB, UTC+7) here; adjust the offset if users
// ever span multiple timezones.
var scheduleLocation = time.FixedZone("WIB", 7*60*60)

func parseScheduledAt(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02T15:04", value, scheduleLocation)
}

// asScheduleLocation re-stamps a time read back from the database into WIB.
// lib/pq labels "timestamp without time zone" columns as UTC on scan
// regardless of what zone was used to write them, so without this,
// comparisons against time.Now() end up off by the WIB offset (7h).
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

	now := time.Now()
	borrow := models.Borrow{
		ID:                      utils.GenerateID(),
		ScheduledDatetime:       scheduled,
		BookedDatetime:          now,
		Reason:                  req.Reason,
		BorrowDurationSeconds:   totalSeconds,
		BorrowStatusID:          statusPending,
		UserID:                  userID,
		FinishedAt:              nil,
		Title:                   req.Title,
		StatusMessage:           "Request is pending for approval.",
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