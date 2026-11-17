package model

import (
	"time"
)

// Equipment merepresentasikan barang/alat milik organisasi yang siap dipinjam
type Equipment struct {
	ID                uint         `gorm:"primaryKey" json:"id"`
	OrganizationID    uint         `gorm:"not null;index" json:"organization_id"`
	Organization      Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organization,omitempty"`
	
	Code              string       `gorm:"size:50;index" json:"code"`           // Kode unik alat (misal: "LAB-LP-01")
	Name              string       `gorm:"not null;size:100" json:"name"`
	Category          string       `gorm:"size:50;default:'Umum'" json:"category"`
	Description       string       `gorm:"type:text" json:"description"`
	
	// Manajemen Kuantitas Stok (Mencegah Duplikasi)
	TotalQuantity     int          `gorm:"default:1" json:"total_quantity"`
	AvailableQuantity int          `gorm:"default:1" json:"available_quantity"`
	
	// Kondisi Fisik Alat: BAIK, RUSAK, HILANG
	Condition         string       `gorm:"default:'BAIK';size:30" json:"condition"`
	Status            string       `gorm:"default:'AVAILABLE';size:20" json:"status"` // AVAILABLE, BORROWED, MAINTENANCE
	
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}