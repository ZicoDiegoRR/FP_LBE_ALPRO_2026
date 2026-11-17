package model

import (
	"time"
)

// UserKeyPair menyimpan pasangan kunci kriptografi asimetris pengguna
type UserKeyPair struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	UserID              uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User                User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	PublicKeyPEM        string    `gorm:"type:text;not null" json:"public_key"`
	EncryptedPrivateKey string    `gorm:"type:text;not null" json:"-"` // Disimpan dalam format PEM (dapat dienkripsi dengan password user)
	CreatedAt           time.Time `json:"created_at"`
}

// VaultFile merepresentasikan dokumen pribadi terenkripsi di dalam brankas digital
type VaultFile struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	OwnerID          uint       `gorm:"not null;index" json:"owner_id"`
	Owner            User       `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"owner,omitempty"`
	Filename         string     `gorm:"not null;size:255" json:"filename"`
	EncryptedPath    string     `gorm:"not null;size:255" json:"-"` // Path fisik di disk server (.enc)
	CryptoAlgorithm  string     `gorm:"not null;size:10" json:"crypto_algorithm"` // AES, DES, RC4
	EncryptedKeyHex  string     `gorm:"type:text;not null" json:"-"` // Kunci simetris terenkripsi
	
	// Integritas Dokumen & Digital Signature (SHA-256 + RSA Signature)
	FileHash         string     `gorm:"size:64;not null" json:"file_hash"` // SHA-256 Digest
	IsSigned         bool       `gorm:"default:false" json:"is_signed"`
	SignatureHex     string     `gorm:"type:text" json:"signature,omitempty"`
	SignedAt         *time.Time `json:"signed_at,omitempty"`
	
	CreatedAt        time.Time  `json:"created_at"`
}

// FileShareRequest mengelola izin akses dan distribusi kunci file terenkripsi antar-pengguna
type FileShareRequest struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	FileID             uint       `gorm:"not null;index" json:"file_id"`
	File               VaultFile  `gorm:"foreignKey:FileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"file,omitempty"`
	RequesterID        uint       `gorm:"not null;index" json:"requester_id"` // Pengguna yang meminta akses
	Requester          User       `gorm:"foreignKey:RequesterID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"requester,omitempty"`
	OwnerID            uint       `gorm:"not null;index" json:"owner_id"`     // Pemilik file
	Owner              User       `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"owner,omitempty"`
	
	Status             string     `gorm:"default:'PENDING';size:20" json:"status"` // PENDING, APPROVED, REJECTED
	EncryptedKeyForReq string     `gorm:"type:text" json:"-"`                      // Kunci simetris file yang dienkripsi dengan Public Key Requester
	Reason             string     `gorm:"size:255" json:"reason"`
	
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}