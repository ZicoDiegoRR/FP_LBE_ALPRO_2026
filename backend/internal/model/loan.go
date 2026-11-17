package model

import (
	"time"
)

// Loan merepresentasikan riwayat peminjaman barang oleh anggota yang sudah disetujui
type Loan struct {
	ID               uint         `gorm:"primaryKey" json:"id"`
	UserID           uint         `gorm:"not null;index" json:"user_id"`
	User             User         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	OrganizationID   uint         `gorm:"not null;index" json:"organization_id"`
	Organization     Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organization,omitempty"`
	EquipmentID      uint         `gorm:"not null;index" json:"equipment_id"`
	Equipment        Equipment    `gorm:"foreignKey:EquipmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"equipment,omitempty"`
	
	Purpose          string       `gorm:"type:text" json:"purpose"`
	Status           string       `gorm:"default:'PENDING';size:20" json:"status"` // PENDING, APPROVED, REJECTED, RETURNED
	
	// Dokumen Persyaratan Terenkripsi (Zero-Plaintext at Rest)
	OriginalFilename string       `gorm:"size:255" json:"original_filename"`
	EncryptedPath    string       `gorm:"size:255" json:"encrypted_path"`
	CryptoAlgorithm  string       `gorm:"size:10" json:"crypto_algorithm"` // AES, DES, RC4
	
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}