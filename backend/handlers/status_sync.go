package handlers

import (
	"log"
	"time"

	"pmo-backend/db"
	"pmo-backend/models"
)

// Equipment statuses relevant to auto-transitions (separate from the
// Borrow statuses above).
const (
	equipStatusAvailable = "JK7QYOKLJO6CEOFG4UVV"
	equipStatusInUse     = "6OEF7PYAFQJZALEETSAW"
)

// recomputeStatus applies the two automatic transition rules to a single
// Borrow, mutating it in place. Returns true if the status actually changed
// (caller is responsible for saving).
//
// Rules:
//  1. Accepted -> Ongoing, once now is between scheduledDatetime and
//     scheduledDatetime+duration. Accepted -> Finished is NOT automatic —
//     that's an admin action, so once the window passes it just stays
//     Accepted (or Ongoing, if it already flipped) until an admin marks it.
//  2. Pending -> Failed, once now passes scheduledDatetime without ever
//     having been Accepted. This is distinct from Cancelled, which only
//     happens when the user intentionally cancels.
//
// Equipment availability is handled separately by syncEquipmentAvailability
// below, since it depends on Accepted (not just Ongoing) and Accepted is
// itself an admin-set status this function never produces.
func recomputeStatus(b *models.Borrow) bool {
	now := time.Now()
	scheduled := asScheduleLocation(b.ScheduledDatetime)

	switch b.BorrowStatusID {
	case statusAccepted:
		endTime := scheduled.Add(time.Duration(b.BorrowDurationSeconds) * time.Second)
		if !now.Before(scheduled) && now.Before(endTime) {
			b.BorrowStatusID = statusOngoing
			b.StatusMessageUpdateTime = now
			return true
		}

	case statusPending:
		if now.After(scheduled) {
			b.BorrowStatusID = statusFailed
			b.StatusMessage = "Automatically marked as failed: request was not approved before the scheduled time."
			b.StatusMessageUpdateTime = now
			return true
		}
	}

	return false
}

// syncEquipmentAvailability enforces the invariant "equipment is In Use if
// and only if it's linked to a borrow that is currently Accepted or
// Ongoing" — Accepted counts too, not just Ongoing, since equipment is
// considered booked the moment it's approved for a slot (even one a month
// out), not just once actively being used. Both Accepted and Ongoing are
// admin-driven statuses this app never sets itself (direct DB edits), so
// there's no transition to hook — this just re-derives the correct state
// from scratch every time it runs.
func syncEquipmentAvailability() {
	var activeBorrows []models.Borrow
	db.DB.Where("borrowstatus_id IN ?", []string{statusAccepted, statusOngoing}).Find(&activeBorrows)

	activeBorrowIDs := make([]string, 0, len(activeBorrows))
	for _, b := range activeBorrows {
		activeBorrowIDs = append(activeBorrowIDs, b.ID)
	}

	shouldBeInUse := make(map[string]bool)
	if len(activeBorrowIDs) > 0 {
		var links []models.EquipmentBorrowItem
		db.DB.Where("borrow_id IN ?", activeBorrowIDs).Find(&links)
		for _, l := range links {
			shouldBeInUse[l.LabEquipmentID] = true
		}
	}

	inUseIDs := make([]string, 0, len(shouldBeInUse))
	for id := range shouldBeInUse {
		inUseIDs = append(inUseIDs, id)
	}
	if len(inUseIDs) > 0 {
		db.DB.Model(&models.LabEquipment{}).
			Where("id IN ? AND equipmentstatus_id != ?", inUseIDs, equipStatusInUse).
			Update("equipmentstatus_id", equipStatusInUse)
	}

	// Everything else currently marked In Use, but not backed by an
	// Accepted/Ongoing borrow, gets released back to Available.
	query := db.DB.Model(&models.LabEquipment{}).Where("equipmentstatus_id = ?", equipStatusInUse)
	if len(inUseIDs) > 0 {
		query = query.Where("id NOT IN ?", inUseIDs)
	}
	query.Update("equipmentstatus_id", equipStatusAvailable)
}

// syncBorrowStatuses re-checks every Accepted/Pending borrow belonging to a
// user and saves any that just transitioned. Called on page load (dashboard,
// profile) so whoever's actively looking always sees fresh data, even
// between the background scheduler's ticks below.
func syncBorrowStatuses(userID string) {
	var borrows []models.Borrow
	db.DB.Where("user_id = ? AND borrowstatus_id IN ?", userID, []string{statusAccepted, statusPending}).
		Find(&borrows)

	for i := range borrows {
		if recomputeStatus(&borrows[i]) {
			db.DB.Save(&borrows[i])
		}
	}
}

// syncAllBorrowStatuses does the same thing, but across every user — this is
// what the background scheduler calls, since it has no single "current
// user" to scope to.
func syncAllBorrowStatuses() {
	var borrows []models.Borrow
	db.DB.Where("borrowstatus_id IN ?", []string{statusAccepted, statusPending}).
		Find(&borrows)

	changed := 0
	for i := range borrows {
		if recomputeStatus(&borrows[i]) {
			db.DB.Save(&borrows[i])
			changed++
		}
	}
	if changed > 0 {
		log.Printf("status sync: updated %d borrow(s)", changed)
	}

	syncEquipmentAvailability()
}

// StartStatusSyncScheduler runs syncAllBorrowStatuses on a fixed interval,
// forever, in the background. Call this once at startup with `go`.
//
// The interval is a trade-off: short enough that statuses don't feel stale,
// long enough that it doesn't add noticeable load. A borrow's schedule is
// entered with minute precision (datetime-local has no seconds), so there's
// no point checking more often than that — every 30s comfortably beats that
// resolution without hammering the database.
func StartStatusSyncScheduler(interval time.Duration) {
	syncAllBorrowStatuses() // run once immediately, don't wait for the first tick

	ticker := time.NewTicker(interval)
	for range ticker.C {
		syncAllBorrowStatuses()
	}
}