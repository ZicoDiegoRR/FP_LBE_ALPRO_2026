package model

import (
	"time"
)

// Organization merepresentasikan entitas organisasi/penyedia barang yang dikelola Admin
type Organization struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Name               string    `gorm:"not null;size:100" json:"name"`
	RedeemCode         string    `gorm:"uniqueIndex;not null;size:10" json:"redeem_code"` // Kode unik untuk peminjam bergabung
	OwnerID            uint      `gorm:"not null;index" json:"owner_id"`                  // User dengan role 'admin'
	Owner              User      `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"owner,omitempty"`
	Description        string    `gorm:"type:text" json:"description"`
	
	// Fitur Kustomisasi Organisasi (Privacy by Design)
	CustomFieldsSchema string    `gorm:"type:text" json:"custom_fields_schema"` // Definisi form dinamis JSON: [{"key":"no_rt","label":"No RT","required":true}]
	RequireDocument    bool      `gorm:"default:false" json:"require_document"` // Apakah peminjaman wajib melampirkan berkas
	DocumentType       string    `gorm:"size:100" json:"document_type"`         // Jenis berkas syarat (misal: "Proposal Resmi", "KTM", "Surat Izin")
	
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// OrganizationMember merepresentasikan pengajuan dan status keanggotaan peminjam
type OrganizationMember struct {
	ID                uint         `gorm:"primaryKey" json:"id"`
	UserID            uint         `gorm:"not null;uniqueIndex:idx_user_org" json:"user_id"`
	User              User         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	OrganizationID    uint         `gorm:"not null;uniqueIndex:idx_user_org" json:"organization_id"`
	Organization      Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organization,omitempty"`
	
	// Biodata Fleksibel (Data Minimization)
	Fullname          string       `gorm:"not null;size:100" json:"fullname"`
	CustomBiodataJSON string       `gorm:"type:text" json:"custom_biodata_json"` // Jawaban form dinamis anggota dalam format JSON
	
	// Status Verifikasi oleh Admin
	Status            string       `gorm:"default:'PENDING';size:20" json:"status"` // PENDING, APPROVED, REJECTED
	
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}