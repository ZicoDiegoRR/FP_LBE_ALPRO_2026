package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"pmo-backend/internal/model"
)

type LoanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

// CreateLoan membuat permohonan peminjaman baru (stok fisik belum dikurangi sebelum disetujui)
func (r *LoanRepository) CreateLoan(loan *model.Loan) error {
	// Validasi ketersediaan stok
	var eq model.Equipment
	if err := r.db.First(&eq, loan.EquipmentID).Error; err != nil {
		return errors.New("barang tidak ditemukan")
	}

	if eq.AvailableQuantity < loan.Quantity {
		return fmt.Errorf("stok tidak mencukupi: tersedia %d unit, diajukan %d unit", eq.AvailableQuantity, loan.Quantity)
	}

	return r.db.Create(loan).Error
}

func (r *LoanRepository) GetLoansByUserID(userID uint) ([]model.Loan, error) {
	var loans []model.Loan
	err := r.db.Preload("Equipment").Preload("Organization").Where("user_id = ?", userID).Order("created_at desc").Find(&loans).Error
	return loans, err
}

func (r *LoanRepository) GetLoansByOrgID(orgID uint) ([]model.Loan, error) {
	var loans []model.Loan
	err := r.db.Preload("User").Preload("Equipment").Where("organization_id = ?", orgID).Order("created_at desc").Find(&loans).Error
	return loans, err
}

func (r *LoanRepository) GetLoanByID(id uint) (*model.Loan, error) {
	var loan model.Loan
	err := r.db.Preload("User").Preload("Equipment").Preload("Organization").First(&loan, id).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *LoanRepository) GetLoanByTransactionCode(code string) (*model.Loan, error) {
	var loan model.Loan
	err := r.db.Preload("User").Preload("Equipment").Preload("Organization").Where("transaction_code = ?", code).First(&loan, code).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

// ApproveAndAllocateStock menyetujui peminjaman dan memotong stok tersedia secara atomik (ACID)
func (r *LoanRepository) ApproveAndAllocateStock(loanID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var loan model.Loan
		if err := tx.First(&loan, loanID).Error; err != nil {
			return err
		}

		if loan.Status != model.LoanStatusPending {
			return errors.New("hanya pengajuan bertatus PENDING yang dapat disetujui")
		}

		// Ambil data alat dengan kueri kunci
		var eq model.Equipment
		if err := tx.First(&eq, loan.EquipmentID).Error; err != nil {
			return err
		}

		if eq.AvailableQuantity < loan.Quantity {
			return errors.New("gagal menyetujui: stok barang saat ini sudah tidak mencukupi")
		}

		// Kurangi stok tersedia
		eq.AvailableQuantity -= loan.Quantity
		if err := tx.Save(&eq).Error; err != nil {
			return err
		}

		// Perbarui status peminjaman menjadi BORROWED (barang diserahkan)
		loan.Status = model.LoanStatusBorrowed
		return tx.Save(&loan).Error
	})
}

// RejectLoan menolak pengajuan peminjaman
func (r *LoanRepository) RejectLoan(loanID uint, reason string) error {
	return r.db.Model(&model.Loan{}).Where("id = ?", loanID).Updates(map[string]interface{}{
		"status":        model.LoanStatusRejected,
		"reject_reason": reason,
	}).Error
}

// RequestReturn (Oleh Peminjam) mengajukan konfirmasi bahwa barang siap/sudah dikembalikan
func (r *LoanRepository) RequestReturn(loanID uint, userID uint) error {
	return r.db.Model(&model.Loan{}).
		Where("id = ? AND user_id = ? AND status = ?", loanID, userID, model.LoanStatusBorrowed).
		Update("status", model.LoanStatusReturnRequested).Error
}

// ConfirmReturn (Oleh Penyedia) memverifikasi kondisi fisik barang dan mengembalikan stok secara atomik
func (r *LoanRepository) ConfirmReturn(loanID uint, condition, notes string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var loan model.Loan
		if err := tx.First(&loan, loanID).Error; err != nil {
			return err
		}

		if loan.Status != model.LoanStatusReturnRequested && loan.Status != model.LoanStatusBorrowed {
			return errors.New("transaksi tidak dalam status pengembalian yang valid")
		}

		now := time.Now()
		loan.Status = model.LoanStatusReturned
		loan.ReturnedDate = &now
		loan.ReturnCondition = condition
		loan.ReturnNotes = notes

		if err := tx.Save(&loan).Error; err != nil {
			return err
		}

		// Hanya kembalikan stok tersedia jika kondisi barang BAIK
		if condition == "BAIK" {
			var eq model.Equipment
			if err := tx.First(&eq, loan.EquipmentID).Error; err != nil {
				return err
			}
			eq.AvailableQuantity += loan.Quantity
			if eq.AvailableQuantity > eq.TotalQuantity {
				eq.AvailableQuantity = eq.TotalQuantity
			}
			return tx.Save(&eq).Error
		}

		return nil
	})
}