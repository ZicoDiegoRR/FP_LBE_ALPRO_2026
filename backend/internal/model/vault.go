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
	EncryptedPrivateKey string    `gorm:"type:text;not null" json:"-"`
	CreatedAt           time.Time `json:"created_at"`
}

// VaultFile merepresentasikan dokumen pribadi terenkripsi di dalam brankas digital
type VaultFile struct {
	ID               uint         `gorm:"primaryKey" json:"id"`
	OwnerID          uint         `gorm:"not null;index" json:"owner_id"`
	Owner            User         `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"owner,omitempty"`
	OrganizationID   *uint        `gorm:"index" json:"organization_id,omitempty"` // Opsional: Terikat ke organisasi
	Organization     *Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"organization,omitempty"`
	
	Filename         string       `gorm:"not null;size:255" json:"filename"`
	EncryptedPath    string       `gorm:"not null;size:255" json:"-"`
	CryptoAlgorithm  string       `gorm:"not null;size:10" json:"crypto_algorithm"`
	EncryptedKeyHex  string       `gorm:"type:text;not null" json:"-"`
	
	// Integritas Dokumen & Tanda Tangan Digital
	FileHash         string       `gorm:"size:64;not null" json:"file_hash"`
	IsSigned         bool         `gorm:"default:false" json:"is_signed"`
	SignatureHex     string       `gorm:"type:text" json:"signature,omitempty"`
	SignedAt         *time.Time   `json:"signed_at,omitempty"`
	
	// Secure Cloud Sharing
	AccessCode       string       `gorm:"size:20;index" json:"access_code,omitempty"` // Kode undangan akses berkas
	IsSharedWithOrg  bool         `gorm:"default:false" json:"is_shared_with_org"`
	
	CreatedAt        time.Time    `json:"created_at"`
}

// FileShareRequest mengelola izin akses dan distribusi kunci file terenkripsi antar-pengguna
type FileShareRequest struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	FileID             uint       `gorm:"not null;index" json:"file_id"`
	File               VaultFile  `gorm:"foreignKey:FileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"file,omitempty"`
	RequesterID        uint       `gorm:"not null;index" json:"requester_id"`
	Requester          User       `gorm:"foreignKey:RequesterID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"requester,omitempty"`
	OwnerID            uint       `gorm:"not null;index" json:"owner_id"`
	Owner              User       `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"owner,omitempty"`
	
	Status             string     `gorm:"default:'PENDING';size:20" json:"status"` // PENDING, APPROVED, REJECTED, REVOKED
	EncryptedKeyForReq string     `gorm:"type:text" json:"-"`
	Reason             string     `gorm:"size:255" json:"reason"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty"` // Waktu pencabutan akses
	
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ActivityLog mencatat jejak audit aktivitas akses dan kriptografi berkas
type ActivityLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Action    string    `gorm:"size:50;not null" json:"action"` // UPLOAD, SIGN, VERIFY, SHARE_REQUEST, SHARE_APPROVE, DOWNLOAD, REVOKE
	Details   string    `gorm:"type:text" json:"details"`
	CreatedAt time.Time `json:"created_at"`
}