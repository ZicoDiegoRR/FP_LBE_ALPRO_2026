package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/config"
	"pmo-backend/internal/crypto"
	"pmo-backend/internal/model"
	"pmo-backend/internal/repository"
)

type LoanHandler struct {
	loanRepo      *repository.LoanRepository
	orgRepo       *repository.OrgRepository
	cryptoService *crypto.CryptoService
	cfg           *config.Config
}

func NewLoanHandler(
	loanRepo *repository.LoanRepository,
	orgRepo *repository.OrgRepository,
	cryptoService *crypto.CryptoService,
	cfg *config.Config,
) *LoanHandler {
	if err := os.MkdirAll(cfg.StoragePath, 0755); err != nil {
		fmt.Printf("[WARN] Gagal membuat direktori storage: %v\n", err)
	}

	return &LoanHandler{
		loanRepo:      loanRepo,
		orgRepo:       orgRepo,
		cryptoService: cryptoService,
		cfg:           cfg,
	}
}

func (h *LoanHandler) getKeyByAlgorithm(algo string) ([]byte, error) {
	cleanAlgo := strings.TrimSpace(strings.ToUpper(algo))
	switch cleanAlgo {
	case "AES":
		return []byte(strings.TrimSpace(h.cfg.CryptoKeyAES)), nil
	case "DES":
		return []byte(strings.TrimSpace(h.cfg.CryptoKeyDES)), nil
	case "RC4":
		return []byte(strings.TrimSpace(h.cfg.CryptoKeyRC4)), nil
	default:
		return nil, fmt.Errorf("algoritma tidak dikenali: %s", algo)
	}
}

func generateTransactionCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	timestamp := time.Now().Format("20060102")
	return fmt.Sprintf("TRX-%s-%s", timestamp, strings.ToUpper(hex.EncodeToString(b)))
}

// CreateLoan mengajukan peminjaman alat dengan validasi durasi & dokumen opsional/wajib
func (h *LoanHandler) CreateLoan(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Verifikasi Keanggotaan Peminjam
	membership, err := h.orgRepo.GetMembershipByUser(userID)
	if err != nil || membership == nil || membership.Status != "APPROVED" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda belum terdaftar atau belum disetujui sebagai anggota organisasi"})
		return
	}

	equipmentIDStr := c.PostForm("equipment_id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Barang yang dipilih tidak valid"})
		return
	}

	qtyStr := c.DefaultPostForm("quantity", "1")
	quantity, _ := strconv.Atoi(qtyStr)
	if quantity < 1 {
		quantity = 1
	}

	purpose := c.PostForm("purpose")
	startDateStr := c.PostForm("start_date")
	dueDateStr := c.PostForm("due_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		startDate = time.Now()
	}
	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		dueDate = startDate.AddDate(0, 0, 3) // Default 3 hari
	}

	// Cek kebijakan dokumen organisasi
	org, _ := h.orgRepo.GetOrgByID(membership.OrganizationID)
	algorithm := c.DefaultPostForm("crypto_algorithm", "AES")

	var origFilename, encPath, fileHash string
	fileHeader, err := c.FormFile("document")

	if org != nil && org.RequireDocument && err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Organisasi mewajibkan lampiran berkas: %s", org.DocumentType)})
		return
	}

	// Jika ada file yang dilampirkan, lakukan enkripsi Zero-Plaintext
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err == nil {
			defer file.Close()
			fileBytes, _ := io.ReadAll(file)

			// Hitung hash
			hBytes := sha256.Sum256(fileBytes)
			fileHash = hex.EncodeToString(hBytes[:])

			key, err := h.getKeyByAlgorithm(algorithm)
			if err == nil {
				encryptedData, err := h.cryptoService.Encrypt(fileBytes, algorithm, key)
				if err == nil {
					timestamp := time.Now().UnixNano()
					safeFilename := fmt.Sprintf("loan_%d_%s.enc", timestamp, filepath.Base(fileHeader.Filename))
					destinationPath := filepath.Join(h.cfg.StoragePath, safeFilename)

					if os.WriteFile(destinationPath, encryptedData, 0600) == nil {
						origFilename = fileHeader.Filename
						encPath = destinationPath
					}
				}
			}
		}
	}

	loan := model.Loan{
		TransactionCode:  generateTransactionCode(),
		UserID:           userID,
		OrganizationID:   membership.OrganizationID,
		EquipmentID:      uint(equipmentID),
		Quantity:         quantity,
		Purpose:          purpose,
		Status:           model.LoanStatusPending,
		StartDate:        startDate,
		DueDate:          dueDate,
		OriginalFilename: origFilename,
		EncryptedPath:    encPath,
		CryptoAlgorithm:  algorithm,
		FileHash:         fileHash,
	}

	if err := h.loanRepo.CreateLoan(&loan); err != nil {
		if encPath != "" {
			os.Remove(encPath)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengajuan peminjaman berhasil dikirim!",
		"loan":    loan,
	})
}

// GetUserLoans mengambil riwayat peminjaman peminjam saat ini
func (h *LoanHandler) GetUserLoans(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	loans, err := h.loanRepo.GetLoansByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar peminjaman"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// GetOrgLoans (Admin) mengambil riwayat seluruh peminjaman di organisasinya
func (h *LoanHandler) GetOrgLoans(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	org, err := h.orgRepo.GetOrgByOwnerID(adminID)
	if err != nil || org == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organisasi tidak ditemukan"})
		return
	}

	loans, err := h.loanRepo.GetLoansByOrgID(org.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar peminjaman"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// ApproveLoan (Admin) menyetujui peminjaman dan mengalokasikan stok secara atomik
func (h *LoanHandler) ApproveLoan(c *gin.Context) {
	loanIDStr := c.Param("id")
	loanID, _ := strconv.ParseUint(loanIDStr, 10, 32)

	if err := h.loanRepo.ApproveAndAllocateStock(uint(loanID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Peminjaman disetujui dan barang resmi diserahkan!"})
}

// RejectLoan (Admin) menolak peminjaman dengan alasan
func (h *LoanHandler) RejectLoan(c *gin.Context) {
	loanIDStr := c.Param("id")
	loanID, _ := strconv.ParseUint(loanIDStr, 10, 32)

	var input struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alasan penolakan wajib diisi"})
		return
	}

	if err := h.loanRepo.RejectLoan(uint(loanID), input.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menolak pengajuan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pengajuan peminjaman ditolak."})
}

// RequestReturn (Peminjam) mengajukan konfirmasi bahwa barang sudah dikembalikan
func (h *LoanHandler) RequestReturn(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	loanIDStr := c.Param("id")
	loanID, _ := strconv.ParseUint(loanIDStr, 10, 32)

	if err := h.loanRepo.RequestReturn(uint(loanID), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal mengajukan pengembalian: pastikan barang sedang dalam status BORROWED"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pengajuan pengembalian dikirim! Menunggu konfirmasi fisik dari pihak penyedia."})
}

// ConfirmReturn (Admin) memverifikasi kondisi fisik barang dan mengembalikan stok
func (h *LoanHandler) ConfirmReturn(c *gin.Context) {
	loanIDStr := c.Param("id")
	loanID, _ := strconv.ParseUint(loanIDStr, 10, 32)

	var input struct {
		Condition string `json:"condition" binding:"required,oneof=BAIK RUSAK HILANG"`
		Notes     string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kondisi barang harus BAIK, RUSAK, atau HILANG"})
		return
	}

	if err := h.loanRepo.ConfirmReturn(uint(loanID), input.Condition, input.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pengembalian berhasil diverifikasi dan transaksi selesai!"})
}

// DownloadDecryptedFile mengunduh berkas dengan Otorisasi Multi-Tenant Ketat
func (h *LoanHandler) DownloadDecryptedFile(c *gin.Context) {
	loanIDStr := c.Param("id")
	loanID, _ := strconv.ParseUint(loanIDStr, 10, 32)

	loan, err := h.loanRepo.GetLoanByID(uint(loanID))
	if err != nil || loan == nil || loan.EncryptedPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	currentUserID := c.MustGet("user_id").(uint)
	currentUserRole := c.MustGet("role").(string)

	// Otorisasi: Pemilik Berkas ATAU Admin yang memiliki organisasi tersebut!
	isOwner := loan.UserID == currentUserID
	isAdminOfOrg := false

	if strings.EqualFold(currentUserRole, "admin") {
		org, _ := h.orgRepo.GetOrgByOwnerID(currentUserID)
		if org != nil && org.ID == loan.OrganizationID {
			isAdminOfOrg = true
		}
	}

	if !isOwner && !isAdminOfOrg {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak memiliki wewenang mengunduh dokumen transaksi ini"})
		return
	}

	encryptedData, err := os.ReadFile(loan.EncryptedPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File fisik tidak ditemukan di server"})
		return
	}

	key, err := h.getKeyByAlgorithm(loan.CryptoAlgorithm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunci enkripsi tidak valid"})
		return
	}

	decryptedPlaintext, err := h.cryptoService.Decrypt(encryptedData, loan.CryptoAlgorithm, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dekripsi gagal: integritas file rusak"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", loan.OriginalFilename))
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", decryptedPlaintext)
}