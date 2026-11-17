package model

import (
	"time"
)

// Organization merepresentasikan entitas organisasi/penyedia barang yang dikelola Admin
type Organization struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;size:100" json:"name"`
	RedeemCode  string    `gorm:"uniqueIndex;not null;size:10" json:"redeem_code"` // Kode unik untuk peminjam bergabung
	OwnerID     uint      `gorm:"not null;index" json:"owner_id"`                  // User dengan role 'admin'
	Owner       User      `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"owner,omitempty"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrganizationMember merepresentasikan pengajuan dan status keanggotaan peminjam
type OrganizationMember struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	UserID         uint         `gorm:"not null;uniqueIndex:idx_user_org" json:"user_id"`
	User           User         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	OrganizationID uint         `gorm:"not null;uniqueIndex:idx_user_org" json:"organization_id"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organization,omitempty"`
	
	// Biodata Peminjam (Lapis Verifikasi)
	Fullname       string       `gorm:"not null;size:100" json:"fullname"`
	IdentityNumber string       `gorm:"not null;size:50" json:"identity_number"` // NIM / NIK
	PhoneNumber    string       `gorm:"not null;size:20" json:"phone_number"`
	Reason         string       `gorm:"type:text" json:"reason"`                 // Alasan bergabung
	
	// Status Verifikasi oleh Admin
	Status         string       `gorm:"default:'PENDING';size:20" json:"status"` // PENDING, APPROVED, REJECTED
	
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}