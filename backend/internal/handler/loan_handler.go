package handler

import (
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

// CreateLoan mengajukan peminjaman alat (dengan enkripsi file syarat)
func (h *LoanHandler) CreateLoan(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Verifikasi Keanggotaan: Peminjam WAJIB disetujui (APPROVED) terlebih dahulu
	membership, err := h.orgRepo.GetMembershipByUser(userID)
	if err != nil || membership == nil || membership.Status != "APPROVED" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Akses ditolak: Anda belum terdaftar atau belum disetujui sebagai anggota organisasi",
		})
		return
	}

	equipmentIDStr := c.PostForm("equipment_id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID alat tidak valid"})
		return
	}

	purpose := c.PostForm("purpose")
	algorithm := c.DefaultPostForm("crypto_algorithm", "AES")

	// Baca file dokumen persyaratan (misal: KTM / Surat)
	fileHeader, err := c.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dokumen pendukung (file) wajib diunggah"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file"})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca byte file"})
		return
	}

	key, err := h.getKeyByAlgorithm(algorithm)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ENKRIPSI IN-MEMORY (Zero-Plaintext at Rest)
	encryptedData, err := h.cryptoService.Encrypt(fileBytes, algorithm, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Enkripsi gagal: " + err.Error()})
		return
	}

	// Simpan ciphertext ke disk
	timestamp := time.Now().UnixNano()
	safeFilename := fmt.Sprintf("%d_%s.enc", timestamp, filepath.Base(fileHeader.Filename))
	destinationPath := filepath.Join(h.cfg.StoragePath, safeFilename)

	if err := os.WriteFile(destinationPath, encryptedData, 0600); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file terenkripsi"})
		return
	}

	loan := model.Loan{
		UserID:           userID,
		OrganizationID:   membership.OrganizationID,
		EquipmentID:      uint(equipmentID),
		Purpose:          purpose,
		Status:           "PENDING",
		OriginalFilename: fileHeader.Filename,
		EncryptedPath:    destinationPath,
		CryptoAlgorithm:  algorithm,
	}

	if err := h.loanRepo.CreateLoan(&loan); err != nil {
		os.Remove(destinationPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengajuan peminjaman"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengajuan peminjaman berhasil dan dokumen telah diamankan!",
		"loan":    loan,
	})
}

// GetUserLoans mengambil riwayat peminjaman milik peminjam saat ini
func (h *LoanHandler) GetUserLoans(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	loans, err := h.loanRepo.GetLoansByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar peminjaman"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// DownloadDecryptedFile mendekripsi file dan mengembalikannya ke user yang berhak
func (h *LoanHandler) DownloadDecryptedFile(c *gin.Context) {
	loanIDStr := c.Param("id")
	loanID, _ := strconv.ParseUint(loanIDStr, 10, 32)

	loan, err := h.loanRepo.GetLoanByID(uint(loanID))
	if err != nil || loan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data peminjaman tidak ditemukan"})
		return
	}

	currentUserID := c.MustGet("user_id").(uint)
	currentUserRole := c.MustGet("role").(string)

	// Pemilik file ATAU Admin organisasi berhak mengunduh
	if loan.UserID != currentUserID && currentUserRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak memiliki wewenang mengunduh berkas ini"})
		return
	}

	encryptedData, err := os.ReadFile(loan.EncryptedPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File terenkripsi tidak ditemukan"})
		return
	}

	key, err := h.getKeyByAlgorithm(loan.CryptoAlgorithm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunci enkripsi tidak valid"})
		return
	}

	decryptedPlaintext, err := h.cryptoService.Decrypt(encryptedData, loan.CryptoAlgorithm, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dekripsi gagal: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", loan.OriginalFilename))
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", decryptedPlaintext)
}