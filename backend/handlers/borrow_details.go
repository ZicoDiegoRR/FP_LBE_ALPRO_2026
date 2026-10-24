package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pmo-backend/db"
	"pmo-backend/models"
)

// labForBorrow implements the same workaround as labNameForBorrow in
// dashboard.go, but returns the full Lab record (needed here for its
// pictureUrl/logo too, not just the name).
func labForBorrow(borrowID string) models.Lab {
	var item models.EquipmentBorrowItem
	if err := db.DB.Where("borrow_id = ?", borrowID).First(&item).Error; err != nil {
		return models.Lab{Name: "Unknown Lab"}
	}

	var labEquipment models.LabEquipment
	if err := db.DB.Where("id = ?", item.LabEquipmentID).First(&labEquipment).Error; err != nil {
		return models.Lab{Name: "Unknown Lab"}
	}

	var lab models.Lab
	if err := db.DB.Where("id = ?", labEquipment.LabID).First(&lab).Error; err != nil {
		return models.Lab{Name: "Unknown Lab"}
	}
	return lab
}

type BorrowDetailEquipmentItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Desc          string `json:"desc"`
	EquipmentType string `json:"equipmentType"`
	PictureURL    string `json:"pictureUrl"`
}

// fetchOwnedBorrow loads a Borrow by id and checks it belongs to the
// logged-in user. Returns (borrow, true) on success.
func fetchOwnedBorrow(c *gin.Context, id string) (models.Borrow, bool) {
	userID := c.GetString("user_id")

	var b models.Borrow
	if err := db.DB.Where("id = ?", id).First(&b).Error; err != nil {
		c.JSON(http.StatusNotFound, APIResponse{false, "Borrow request not found"})
		return b, false
	}
	if b.UserID != userID {
		c.JSON(http.StatusForbidden, APIResponse{false, "This isn't your borrow request"})
		return b, false
	}
	return b, true
}

// GET /borrow/:id (requires auth)
func GetBorrowDetail(c *gin.Context) {
	b, ok := fetchOwnedBorrow(c, c.Param("id"))
	if !ok {
		return
	}

	if recomputeStatus(&b) {
		db.DB.Save(&b)
	}

	lab := labForBorrow(b.ID)

	var finishedAt interface{}
	if b.FinishedAt != nil {
		finishedAt = b.FinishedAt.Format("Jan 2, 2006 15:04")
	} else {
		finishedAt = nil
	}

	// Equipment actually tied to THIS borrow (via EquipmentBorrowItem),
	// not the lab's whole inventory.
	var links []models.EquipmentBorrowItem
	db.DB.Where("borrow_id = ?", b.ID).Find(&links)

	labEquipIDs := make([]string, 0, len(links))
	for _, l := range links {
		labEquipIDs = append(labEquipIDs, l.LabEquipmentID)
	}

	var labEquipRows []models.LabEquipment
	db.DB.Where("id IN ?", labEquipIDs).Find(&labEquipRows)

	equipIDs := make([]string, 0, len(labEquipRows))
	for _, le := range labEquipRows {
		equipIDs = append(equipIDs, le.EquipmentID)
	}
	var equipRows []models.Equipment
	db.DB.Where("id IN ?", equipIDs).Find(&equipRows)
	equipByID := make(map[string]models.Equipment, len(equipRows))
	for _, e := range equipRows {
		equipByID[e.ID] = e
	}

	items := make([]BorrowDetailEquipmentItem, 0, len(labEquipRows))
	for _, le := range labEquipRows {
		eq := equipByID[le.EquipmentID]
		items = append(items, BorrowDetailEquipmentItem{
			ID:            le.ID,
			Name:          le.Name,
			Desc:          le.Desc,
			EquipmentType: eq.Name,
			PictureURL:    eq.PictureURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"borrow": gin.H{
			"id":                      b.ID,
			"title":                   b.Title,
			"bookedAt":                b.BookedDatetime.Format("Jan 2, 2006 15:04"),
			"scheduledAt":             b.ScheduledDatetime.Format("Jan 2, 2006 15:04"),
			"statusMessageUpdateTime": b.StatusMessageUpdateTime.Format("Jan 2, 2006 15:04"),
			"finishedAt":              finishedAt,
			"labName":                 lab.Name,
			"labPictureUrl":           lab.PictureURL,
			"reason":                  b.Reason,
			"statusMessage":           b.StatusMessage,
			"statusType":              statusTypeForID(b.BorrowStatusID),
		},
		"equipment": items,
	})
}

// POST /borrow/:id/cancel (requires auth)
// Only allowed while the request is still Pending.
func CancelBorrow(c *gin.Context) {
	b, ok := fetchOwnedBorrow(c, c.Param("id"))
	if !ok {
		return
	}

	if b.BorrowStatusID != statusPending {
		c.JSON(http.StatusConflict, APIResponse{false, "Only pending requests can be cancelled"})
		return
	}

	b.BorrowStatusID = statusCancelled
	b.StatusMessage = "Cancelled by user."
	b.StatusMessageUpdateTime = time.Now()

	if err := db.DB.Save(&b).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not cancel request"})
		return
	}

	c.JSON(http.StatusOK, APIResponse{true, "Borrow request cancelled"})
}