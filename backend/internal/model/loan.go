package model

import (
	"time"
)

// Loan merepresentasikan permohonan peminjaman dan arsip file terenkripsi
type Loan struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index" json:"user_id"`
	User           User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	ItemName       string    `gorm:"not null;size:100" json:"item_name"`
	Description    string    `gorm:"type:text" json:"description"`
	Status         string    `gorm:"default:'PENDING';size:20" json:"status"` // PENDING, APPROVED, REJECTED
	
	// Data Kriptografi File Pendukung
	OriginalFilename string  `gorm:"size:255" json:"original_filename"`
	EncryptedPath    string  `gorm:"size:255" json:"encrypted_path"`
	CryptoAlgorithm  string  `gorm:"size:10" json:"crypto_algorithm"` // AES, DES, atau RC4
	
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}