package model

import (
	"time"
)

// Equipment merepresentasikan barang/alat milik organisasi yang siap dipinjam
type Equipment struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	OrganizationID uint         `gorm:"not null;index" json:"organization_id"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organization,omitempty"`
	Name           string       `gorm:"not null;size:100" json:"name"`
	Description    string       `gorm:"type:text" json:"description"`
	Status         string       `gorm:"default:'AVAILABLE';size:20" json:"status"` // AVAILABLE, BORROWED, MAINTENANCE
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}