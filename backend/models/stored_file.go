package models

import "time"

// StoredFile adalah metadata berkas terenkripsi milik satu pengguna.
// Isi berkas ada di backend/vault/<ID>.enc (format filecrypt), bukan di database.
type StoredFile struct {
	ID         string `gorm:"primaryKey;size:20"`
	OwnerID    string `gorm:"index;size:20"`
	Category   string // identity_scan | document | video
	OrigName   string
	MimeType   string // hasil deteksi server, bukan dari klien
	Cipher     string // aes | des | rc4
	PlainSize  int64
	CipherSize int64
	CreatedAt  time.Time
}

func (StoredFile) TableName() string {
	return "stored_files"
}
