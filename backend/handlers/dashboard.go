package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"pmo-backend/db"
	"pmo-backend/encrypt"
	"pmo-backend/models"
)

// Borrow status IDs (from the seeded BorrowStatus table).
const (
	statusAccepted  = "QTZ9IYVBZTAMKEIAEUZL"
	statusRejected  = "W0ABU8J8NTBB7ERDLOF5"
	statusCancelled = "ITNNA5NGDAJ1JMBJ0VZJ"
	statusFailed    = "AC0TBLTYQUYCSJ096T7R"
	statusPending   = "8371HA2RZB8EKYO7FU6T"
	statusFinished  = "IKXUHNQBZG523YCGUB7J"
	statusOngoing   = "0KHQJHJSO3DUHGU53IY5"
)

const activityTextLimit = 60

// truncate shortens s to at most n characters, adding "..." if it was cut.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// formatDuration turns a second count into something like "2h 15m".
func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if h > 0 && m > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	} else if h > 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", m)
}

// labNameForBorrow implements the workaround: Borrow has no Lab_id directly,
// so we go Borrow -> EquipmentBorrowItem (first row) -> LabEquipment -> Lab.
// All equipment on one borrow is guaranteed to share the same lab.
func labNameForBorrow(borrowID string) string {
	var item models.EquipmentBorrowItem
	if err := db.DB.Where("borrow_id = ?", borrowID).First(&item).Error; err != nil {
		return "Unknown Lab"
	}

	var labEquipment models.LabEquipment
	if err := db.DB.Where("id = ?", item.LabEquipmentID).First(&labEquipment).Error; err != nil {
		return "Unknown Lab"
	}

	var lab models.Lab
	if err := db.DB.Where("id = ?", labEquipment.LabID).First(&lab).Error; err != nil {
		return "Unknown Lab"
	}

	return lab.Name
}

type ActivityItem struct {
	StatusType string `json:"statusType"` // "cancelled" | "pending" | "accepted" | "rejected" | "ongoing" | "finished" | "none"
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
}

func buildActivityItem(b models.Borrow) ActivityItem {
	lab := labNameForBorrow(b.ID)

	key := encrypt.GetEncryptionKey()
	algo := encrypt.GetAlgorithm()

	// Decrypt Reason and StatusMessage for display
	decReason, err := encrypt.DecryptMiddleware(b.Reason, key, algo)
	if err != nil {
		decReason = b.Reason
	}

	decStatusMsg, err := encrypt.DecryptMiddleware(b.StatusMessage, key, algo)
	if err != nil {
		decStatusMsg = b.StatusMessage
	}

	switch b.BorrowStatusID {
	case statusCancelled:
		return ActivityItem{
			StatusType: "cancelled",
			Line1:      fmt.Sprintf("Cancelled borrow request to %s scheduled at %s.", lab, b.ScheduledDatetime.Format("Jan 2, 2006 15:04")),
			Line2:      "Reason: " + truncate(decStatusMsg, activityTextLimit),
		}
	case statusPending:
		return ActivityItem{
			StatusType: "pending",
			Line1:      fmt.Sprintf("Borrow request form submitted to %s. Wait for approval.", lab),
			Line2:      "Reason: " + truncate(decReason, activityTextLimit),
		}
	case statusAccepted:
		return ActivityItem{
			StatusType: "accepted",
			Line1:      fmt.Sprintf("Submission to %s was accepted.", lab),
			Line2:      "Scheduled at: " + b.ScheduledDatetime.Format("Jan 2, 2006 15:04"),
		}
	case statusRejected:
		return ActivityItem{
			StatusType: "rejected",
			Line1:      fmt.Sprintf("Submission to %s was rejected.", lab),
			Line2:      "Reason: " + truncate(decStatusMsg, activityTextLimit),
		}
	case statusOngoing:
		return ActivityItem{
			StatusType: "ongoing",
			Line1:      fmt.Sprintf("Borrowing schedule at %s is ongoing.", lab),
			Line2:      "Active for: " + formatDuration(b.BorrowDurationSeconds),
		}
	case statusFinished:
		return ActivityItem{
			StatusType: "finished",
			Line1:      fmt.Sprintf("Borrowing at %s has finished.", lab),
		}
	case statusFailed:
		return ActivityItem{
			StatusType: "failed",
			Line1:      fmt.Sprintf("Borrow request to %s failed.", lab),
			Line2:      "Reason: " + truncate(decStatusMsg, activityTextLimit),
		}
	default:
		return ActivityItem{
			StatusType: "none",
			Line1:      truncate(decStatusMsg, activityTextLimit),
		}
	}
}

// GET /dashboard (requires auth)
func GetDashboard(c *gin.Context) {
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")

	syncBorrowStatuses(userID)

	var totalBorrowings, ongoing, waitingForApproval, cancelledRejected int64

	db.DB.Model(&models.Borrow{}).Where("user_id = ?", userID).Count(&totalBorrowings)
	db.DB.Model(&models.Borrow{}).Where("user_id = ? AND borrowstatus_id = ?", userID, statusOngoing).Count(&ongoing)
	db.DB.Model(&models.Borrow{}).Where("user_id = ? AND borrowstatus_id = ?", userID, statusPending).Count(&waitingForApproval)
	db.DB.Model(&models.Borrow{}).
		Where("user_id = ? AND borrowstatus_id IN ?", userID, []string{statusCancelled, statusRejected}).
		Count(&cancelledRejected)

	var recentBorrows []models.Borrow
	db.DB.Where("user_id = ?", userID).
		Order("statusmessageupdatetime DESC").
		Limit(5).
		Find(&recentBorrows)

	activity := make([]ActivityItem, 0, len(recentBorrows))
	for _, b := range recentBorrows {
		activity = append(activity, buildActivityItem(b))
	}
	if len(activity) == 0 {
		activity = append(activity, ActivityItem{StatusType: "none", Line1: "No recent activities yet."})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    gin.H{"name": userName},
		"stats": gin.H{
			"totalBorrowings":    totalBorrowings,
			"ongoing":            ongoing,
			"waitingForApproval": waitingForApproval,
			"cancelledRejected":  cancelledRejected,
		},
		"recentActivity": activity,
	})
}