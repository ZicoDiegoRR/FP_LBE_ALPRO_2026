package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pmo-backend/db"
	"pmo-backend/models"
)

type LabOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GET /labs/:id/equipment
// Returns everything the "Select Equipment" screen needs in one call: the
// lab's own info (for its logo), plus each LabEquipment row joined with its
// Equipment (type/picture) and EquipmentStatus (name/symbol).
type EquipmentListItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`          // LabEquipment.Name (this specific item)
	Desc          string `json:"desc"`          // LabEquipment.Desc
	EquipmentType string `json:"equipmentType"` // Equipment.Name
	PictureURL    string `json:"pictureUrl"`    // Equipment.PictureURL
	StatusID      string `json:"statusId"`
	StatusName    string `json:"statusName"`
	StatusSymbol  string `json:"statusSymbol"`
}

func GetLabEquipment(c *gin.Context) {
	labID := c.Param("id")

	syncEquipmentAvailability()

	var lab models.Lab
	if err := db.DB.Where("id = ?", labID).First(&lab).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Lab not found"})
		return
	}

	var labEquipment []models.LabEquipment
	db.DB.Where("lab_id = ?", labID).Find(&labEquipment)

	// Batch-fetch the Equipment and EquipmentStatus rows referenced, rather
	// than querying once per item.
	equipmentIDs := make([]string, 0, len(labEquipment))
	statusIDs := make([]string, 0, len(labEquipment))
	for _, le := range labEquipment {
		equipmentIDs = append(equipmentIDs, le.EquipmentID)
		statusIDs = append(statusIDs, le.EquipmentStatusID)
	}

	var equipmentRows []models.Equipment
	db.DB.Where("id IN ?", equipmentIDs).Find(&equipmentRows)
	equipmentByID := make(map[string]models.Equipment, len(equipmentRows))
	for _, e := range equipmentRows {
		equipmentByID[e.ID] = e
	}

	var statusRows []models.EquipmentStatus
	db.DB.Where("id IN ?", statusIDs).Find(&statusRows)
	statusByID := make(map[string]models.EquipmentStatus, len(statusRows))
	for _, s := range statusRows {
		statusByID[s.ID] = s
	}

	items := make([]EquipmentListItem, 0, len(labEquipment))
	for _, le := range labEquipment {
		eq := equipmentByID[le.EquipmentID]
		st := statusByID[le.EquipmentStatusID]
		items = append(items, EquipmentListItem{
			ID:            le.ID,
			Name:          le.Name,
			Desc:          le.Desc,
			EquipmentType: eq.Name,
			PictureURL:    eq.PictureURL,
			StatusID:      st.ID,
			StatusName:    st.Name,
			StatusSymbol:  st.Symbol,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"lab":     gin.H{"id": lab.ID, "name": lab.Name, "pictureUrl": lab.PictureURL},
		"equipment": items,
	})
}

// GET /labs — used to populate the "Laboratory" dropdown on the borrow form.
func ListLabs(c *gin.Context) {
	var labs []models.Lab
	if err := db.DB.Order("name ASC").Find(&labs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Could not load labs"})
		return
	}

	options := make([]LabOption, 0, len(labs))
	for _, l := range labs {
		options = append(options, LabOption{ID: l.ID, Name: l.Name})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "labs": options})
}