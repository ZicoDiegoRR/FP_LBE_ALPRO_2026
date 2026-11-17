package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/config"
	"pmo-backend/internal/crypto"
	"pmo-backend/internal/model"
	"pmo-backend/internal/repository"

	"strings"
)

// LoanHandler mengelola alur HTTP peminjaman dan arsip file terenkripsi
type LoanHandler struct {
	loanRepo      *repository.LoanRepository
	cryptoService *crypto.CryptoService
	cfg           *config.Config
}

// NewLoanHandler membuat instance baru LoanHandler
func NewLoanHandler(loanRepo *repository.LoanRepository, cryptoService *crypto.CryptoService, cfg *config.Config) *LoanHandler {
	// Pastikan direktori penyimpanan file terenkripsi sudah ada
	if err := os.MkdirAll(cfg.StoragePath, 0755); err != nil {
		fmt.Printf("[WARN] Gagal membuat direktori storage: %v\n", err)
	}

	return &LoanHandler{
		loanRepo:      loanRepo,
		cryptoService: cryptoService,
		cfg:           cfg,
	}
}

// helper untuk mengambil key yang sesuai dengan algoritma dari config
func (h *LoanHandler) getKeyByAlgorithm(algo string) ([]byte, error) {
	cleanAlgo := strings.TrimSpace(strings.ToUpper(algo))

	switch cleanAlgo {
	case "AES":
		key := strings.TrimSpace(h.cfg.CryptoKeyAES)
		return []byte(key), nil
	case "DES":
		key := strings.TrimSpace(h.cfg.CryptoKeyDES)
		return []byte(key), nil
	case "RC4":
		key := strings.TrimSpace(h.cfg.CryptoKeyRC4)
		return []byte(key), nil
	default:
		return nil, fmt.Errorf("algoritma tidak dikenali: %s", algo)
	}
}

// CreateLoan menangani pembuatan pengajuan peminjaman + enkripsi file dokumen
func (h *LoanHandler) CreateLoan(c *gin.Context) {
	// 1. Ambil User ID dari context JWT (diisi oleh AuthMiddleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}
	userID := userIDVal.(uint)

	// 2. Baca form data
	itemName := c.PostForm("item_name")
	description := c.PostForm("description")
	algorithm := c.DefaultPostForm("crypto_algorithm", "AES") // Default: AES

	if itemName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama barang/peminjaman wajib diisi"})
		return
	}

	// 3. Baca file dokumen yang diunggah
	fileHeader, err := c.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dokumen pendukung (file) wajib diunggah"})
		return
	}

	// Batasi ukuran maksimal file (misal: 10 MB) untuk mencegah DoS
	if fileHeader.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran file melebihi batas maksimal 10MB"})
		return
	}

	// Buka stream file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file"})
		return
	}
	defer file.Close()

	// Baca seluruh byte file ke RAM (Plaintext)
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses byte file"})
		return
	}

	// 4. Ambil kunci kriptografi yang sesuai dari server config
	key, err := h.getKeyByAlgorithm(algorithm)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. ENKRIPSI FILE ON-THE-FLY DI MEMORI (Zero Plaintext at Rest)
	encryptedData, err := h.cryptoService.Encrypt(fileBytes, algorithm, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proses enkripsi dokumen gagal: " + err.Error()})
		return
	}

	// 6. Simpan data terenkripsi ke disk dengan nama unik
	timestamp := time.Now().UnixNano()
	safeFilename := fmt.Sprintf("%d_%s.enc", timestamp, filepath.Base(fileHeader.Filename))
	destinationPath := filepath.Join(h.cfg.StoragePath, safeFilename)

	if err := os.WriteFile(destinationPath, encryptedData, 0600); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file terenkripsi ke storage"})
		return
	}

	// 7. Simpan metadata peminjaman ke database
	loan := model.Loan{
		UserID:           userID,
		ItemName:         itemName,
		Description:      description,
		Status:           "PENDING",
		OriginalFilename: fileHeader.Filename,
		EncryptedPath:    destinationPath,
		CryptoAlgorithm:  algorithm,
	}

	if err := h.loanRepo.CreateLoan(&loan); err != nil {
		// Rollback file jika insert database gagal
		os.Remove(destinationPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data pengajuan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengajuan peminjaman berhasil dibuat dan dokumen telah diamankan!",
		"loan":    loan,
	})
}

// GetUserLoans mengambil seluruh daftar peminjaman milik user yang sedang login
func (h *LoanHandler) GetUserLoans(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	loans, err := h.loanRepo.GetLoansByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar peminjaman"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// GetAllLoans mengambil seluruh peminjaman dari semua user (khusus role Admin)
func (h *LoanHandler) GetAllLoans(c *gin.Context) {
	loans, err := h.loanRepo.GetAllLoans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data peminjaman"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// DownloadDecryptedFile membaca file terenkripsi, mendekripsinya kembali, lalu mengirimkannya ke browser
func (h *LoanHandler) DownloadDecryptedFile(c *gin.Context) {
	loanIDStr := c.Param("id")
	loanID, err := strconv.ParseUint(loanIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID peminjaman tidak valid"})
		return
	}

	loan, err := h.loanRepo.GetLoanByID(uint(loanID))
	if err != nil || loan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data peminjaman tidak ditemukan"})
		return
	}

	// Authorization Check: Hanya pemilik dokumen atau Admin yang boleh mengunduh
	currentUserID := c.MustGet("user_id").(uint)
	currentUserRole := c.MustGet("role").(string)

	if loan.UserID != currentUserID && currentUserRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda bukan pemilik file ini"})
		return
	}

	// 1. Baca file ciphertext dari disk
	encryptedData, err := os.ReadFile(loan.EncryptedPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File fisik terenkripsi tidak ditemukan di server"})
		return
	}

	// 2. Ambil kunci kriptografi
	key, err := h.getKeyByAlgorithm(loan.CryptoAlgorithm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunci enkripsi tidak valid"})
		return
	}

	// 3. DEKRIPSI ON-THE-FLY
	decryptedPlaintext, err := h.cryptoService.Decrypt(encryptedData, loan.CryptoAlgorithm, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendekripsi file: integritas data rusak atau kunci salah"})
		return
	}

	// 4. Kirim file plaintext ke browser sebagai attachment download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", loan.OriginalFilename))
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", decryptedPlaintext)
}