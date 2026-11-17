package model

import (
	"time"
)

// Definisi Status Siklus Peminjaman
const (
	LoanStatusPending         = "PENDING"          // Menunggu Persetujuan Admin
	LoanStatusApproved        = "APPROVED"         // Disetujui (Siap diambil)
	LoanStatusBorrowed        = "BORROWED"         // Barang sedang dibawa peminjam
	LoanStatusReturnRequested = "RETURN_REQUESTED" // Peminjam mengajukan pengembalian
	LoanStatusReturned        = "RETURNED"         // Admin telah memverifikasi fisik barang (Selesai)
	LoanStatusRejected        = "REJECTED"         // Ditolak
)

// Loan merepresentasikan transaksi peminjaman lengkap dengan siklus 2-pihak
type Loan struct {
	ID               uint         `gorm:"primaryKey" json:"id"`
	TransactionCode  string       `gorm:"uniqueIndex;size:30" json:"transaction_code"` // Kode Unik / QR Code
	UserID           uint         `gorm:"not null;index" json:"user_id"`
	User             User         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	OrganizationID   uint         `gorm:"not null;index" json:"organization_id"`
	Organization     Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organization,omitempty"`
	EquipmentID      uint         `gorm:"not null;index" json:"equipment_id"`
	Equipment        Equipment    `gorm:"foreignKey:EquipmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"equipment,omitempty"`
	
	Quantity         int          `gorm:"default:1" json:"quantity"` // Jumlah barang yang dipinjam
	Purpose          string       `gorm:"type:text" json:"purpose"`
	Status           string       `gorm:"default:'PENDING';size:30;index" json:"status"`
	RejectReason     string       `gorm:"type:text" json:"reject_reason,omitempty"`
	
	// Durasi & Batas Waktu
	StartDate        time.Time    `json:"start_date"`
	DueDate          time.Time    `json:"due_date"`
	ReturnedDate     *time.Time   `json:"returned_date,omitempty"`
	
	// Mekanisme Pengembalian 2-Pihak
	ReturnCondition  string       `gorm:"size:30" json:"return_condition,omitempty"` // BAIK, RUSAK, HILANG
	ReturnNotes      string       `gorm:"type:text" json:"return_notes,omitempty"`
	
	// Dokumen Persyaratan & Tanda Tangan Digital
	OriginalFilename string       `gorm:"size:255" json:"original_filename"`
	EncryptedPath    string       `gorm:"size:255" json:"encrypted_path"`
	CryptoAlgorithm  string       `gorm:"size:10" json:"crypto_algorithm"` // AES, DES, RC4
	FileHash         string       `gorm:"size:64" json:"file_hash,omitempty"` // SHA-256 Digest
	IsSigned         bool         `gorm:"default:false" json:"is_signed"`
	SignatureHex     string       `gorm:"type:text" json:"signature,omitempty"`
	
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// IsOverdue menghitung status keterlambatan secara otomatis
func (l *Loan) IsOverdue() bool {
	if l.Status == LoanStatusBorrowed && time.Now().After(l.DueDate) {
		return true
	}
	return false
}