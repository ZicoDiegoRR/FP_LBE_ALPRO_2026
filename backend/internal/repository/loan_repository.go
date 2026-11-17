package repository

import (
	"gorm.io/gorm"
	"pmo-backend/internal/model"
)

// LoanRepository mendefinisikan operasi database untuk data Peminjaman
type LoanRepository struct {
	db *gorm.DB
}

// NewLoanRepository membuat instance baru LoanRepository
func NewLoanRepository(db *gorm.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

// CreateLoan menyimpan data pengajuan peminjaman baru
func (r *LoanRepository) CreateLoan(loan *model.Loan) error {
	return r.db.Create(loan).Error
}

// GetLoansByUserID mengambil seluruh daftar peminjaman milik user tertentu
func (r *LoanRepository) GetLoansByUserID(userID uint) ([]model.Loan, error) {
	var loans []model.Loan
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&loans).Error
	return loans, err
}

// GetAllLoans mengambil seluruh peminjaman dari semua user (khusus Admin)
func (r *LoanRepository) GetAllLoans() ([]model.Loan, error) {
	var loans []model.Loan
	// Preload "User" agar data username peminjam ikut terambil (JOIN aman)
	err := r.db.Preload("User").Order("created_at desc").Find(&loans).Error
	return loans, err
}

// GetLoanByID mencari peminjaman spesifik berdasarkan ID
func (r *LoanRepository) GetLoanByID(id uint) (*model.Loan, error) {
	var loan model.Loan
	err := r.db.Preload("User").First(&loan, id).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

// UpdateStatus memperbarui status peminjaman (APPROVED / REJECTED)
func (r *LoanRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Loan{}).Where("id = ?", id).Update("status", status).Error
}