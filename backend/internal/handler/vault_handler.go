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
	"time"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/config"
	"pmo-backend/internal/crypto"
	"pmo-backend/internal/model"
	"pmo-backend/internal/repository"
)

type VaultHandler struct {
	vaultRepo     *repository.VaultRepository
	cryptoService *crypto.CryptoService
	cfg           *config.Config
}

func NewVaultHandler(
	vaultRepo *repository.VaultRepository,
	cryptoService *crypto.CryptoService,
	cfg *config.Config,
) *VaultHandler {
	return &VaultHandler{
		vaultRepo:     vaultRepo,
		cryptoService: cryptoService,
		cfg:           cfg,
	}
}

// 1. GenerateKeyPair: Membuat pasangan kunci RSA untuk user saat ini
func (h *VaultHandler) GenerateKeyPair(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Cek apakah user sudah punya kunci
	existing, _ := h.vaultRepo.GetKeyPairByUserID(userID)
	if existing != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Kunci RSA Anda sudah aktif",
			"public_key": existing.PublicKeyPEM,
		})
		return
	}

	pubPEM, privPEM, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghasilkan kunci RSA: " + err.Error()})
		return
	}

	keyPair := model.UserKeyPair{
		UserID:              userID,
		PublicKeyPEM:        pubPEM,
		EncryptedPrivateKey: privPEM,
	}

	if err := h.vaultRepo.SaveKeyPair(&keyPair); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan kunci RSA"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Pasangan kunci RSA 2048-bit berhasil dibuat!",
		"public_key": pubPEM,
	})
}

// 2. UploadVaultFile: Mengunggah file ke brankas rahasia (AES-256-GCM + SHA-256 Hash)
func (h *VaultHandler) UploadVaultFile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	fileHeader, err := c.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File dokumen wajib diunggah"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses byte file"})
		return
	}

	// Hitung SHA-256 digest untuk integritas
	hashArray := sha256.Sum256(fileBytes)
	fileHashHex := hex.EncodeToString(hashArray[:])

	// Buat kunci simetris AES acak 32-byte khusus untuk file ini (Random Per-File Key)
	perFileKey := make([]byte, 32)
	rand.Read(perFileKey)

	// Enkripsi file in-memory dengan AES-256-GCM
	encryptedBytes, err := crypto.EncryptAES(fileBytes, perFileKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Enkripsi file gagal: " + err.Error()})
		return
	}

	// Simpan ciphertext ke disk
	timestamp := time.Now().UnixNano()
	safeFilename := fmt.Sprintf("vault_%d_%s.enc", timestamp, filepath.Base(fileHeader.Filename))
	destinationPath := filepath.Join(h.cfg.StoragePath, safeFilename)

	if err := os.WriteFile(destinationPath, encryptedBytes, 0600); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file ke disk"})
		return
	}

	vaultFile := model.VaultFile{
		OwnerID:         userID,
		Filename:        fileHeader.Filename,
		EncryptedPath:   destinationPath,
		CryptoAlgorithm: "AES",
		EncryptedKeyHex: hex.EncodeToString(perFileKey),
		FileHash:        fileHashHex,
		IsSigned:        false,
	}

	if err := h.vaultRepo.SaveFile(&vaultFile); err != nil {
		os.Remove(destinationPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencatat metadata file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "File berhasil diamankan di Brankas Pribadi!",
		"file":      vaultFile,
		"file_hash": fileHashHex,
	})
}

// 3. SignDocument: Menandatangani dokumen PDF secara digital dengan RSA Private Key
func (h *VaultHandler) SignDocument(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	fileIDStr := c.Param("id")
	fileID, _ := strconv.ParseUint(fileIDStr, 10, 32)

	file, err := h.vaultRepo.GetFileByID(uint(fileID))
	if err != nil || file == nil || file.OwnerID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan atau bukan milik Anda"})
		return
	}

	// Ambil Private Key user
	keyPair, err := h.vaultRepo.GetKeyPairByUserID(userID)
	if err != nil || keyPair == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Anda belum membuat pasangan kunci RSA"})
		return
	}

	// Baca file terenkripsi dan dekripsi kembali untuk ditandatangani
	encryptedBytes, err := os.ReadFile(file.EncryptedPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file fisik"})
		return
	}

	keyBytes, _ := hex.DecodeString(file.EncryptedKeyHex)
	plainBytes, err := crypto.DecryptAES(encryptedBytes, keyBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendekripsi isi dokumen"})
		return
	}

	// Bubuhkan Digital Signature RSA pada hash dokumen
	sigBytes, err := crypto.SignDocument(plainBytes, keyPair.EncryptedPrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat tanda tangan digital: " + err.Error()})
		return
	}

	sigHex := hex.EncodeToString(sigBytes)
	if err := h.vaultRepo.UpdateSignature(file.ID, sigHex); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status tanda tangan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Dokumen resmi ditandatangani secara digital!",
		"signature": sigHex,
		"file_hash": file.FileHash,
	})
}

// 4. VerifySignature: Memverifikasi keabsahan tanda tangan dokumen
func (h *VaultHandler) VerifySignature(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, _ := strconv.ParseUint(fileIDStr, 10, 32)

	file, err := h.vaultRepo.GetFileByID(uint(fileID))
	if err != nil || file == nil || !file.IsSigned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dokumen belum memiliki tanda tangan digital"})
		return
	}

	// Ambil Public Key pemilik
	pubKeyPEM, err := h.vaultRepo.GetPublicKeyByUserID(file.OwnerID)
	if err != nil || pubKeyPEM == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public key pemilik dokumen tidak ditemukan"})
		return
	}

	encryptedBytes, _ := os.ReadFile(file.EncryptedPath)
	keyBytes, _ := hex.DecodeString(file.EncryptedKeyHex)
	plainBytes, _ := crypto.DecryptAES(encryptedBytes, keyBytes)

	sigBytes, _ := hex.DecodeString(file.SignatureHex)

	// Verifikasi tanda tangan terhadap isi file menggunakan Public Key
	err = crypto.VerifyDocumentSignature(plainBytes, sigBytes, pubKeyPEM)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":   "INVALID",
			"valid":    false,
			"message":  "PERINGATAN: Tanda tangan TIDAK VALID atau dokumen telah dimanipulasi!",
			"filename": file.Filename,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "VALID",
		"valid":     true,
		"message":   "Tanda tangan digital SAH dan dokumen terbukti ASLI (Belum pernah diubah)",
		"signer":    file.Owner.Username,
		"signed_at": file.SignedAt,
		"file_hash": file.FileHash,
	})
}

// 5. RequestAccess: Peminjam meminta akses ke dokumen terenkripsi milik orang lain
func (h *VaultHandler) RequestAccess(c *gin.Context) {
	requesterID := c.MustGet("user_id").(uint)

	var input struct {
		FileID uint   `json:"file_id" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID berkas dan alasan permohonan wajib diisi"})
		return
	}

	file, err := h.vaultRepo.GetFileByID(input.FileID)
	if err != nil || file == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Berkas tidak ditemukan"})
		return
	}

	if file.OwnerID == requesterID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Anda adalah pemilik berkas ini"})
		return
	}

	req := model.FileShareRequest{
		FileID:      file.ID,
		RequesterID: requesterID,
		OwnerID:     file.OwnerID,
		Reason:      input.Reason,
		Status:      "PENDING",
	}

	if err := h.vaultRepo.CreateShareRequest(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengajukan permohonan akses"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Permohonan akses berkas berhasil dikirim ke pemilik!",
		"request": req,
	})
}

// 6. ApproveShareRequest: Pemilik menyetujui -> Enkripsi kunci file dengan Public Key Requester (Hybrid Encryption)
func (h *VaultHandler) ApproveShareRequest(c *gin.Context) {
	ownerID := c.MustGet("user_id").(uint)
	reqIDStr := c.Param("id")
	reqID, _ := strconv.ParseUint(reqIDStr, 10, 32)

	req, err := h.vaultRepo.GetShareRequestByID(uint(reqID))
	if err != nil || req == nil || req.OwnerID != ownerID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permohonan tidak ditemukan atau bukan milik Anda"})
		return
	}

	// 1. Ambil Public Key Requester
	requesterPubKeyPEM, err := h.vaultRepo.GetPublicKeyByUserID(req.RequesterID)
	if err != nil || requesterPubKeyPEM == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pemohon belum menghasilkan kunci publik RSA"})
		return
	}

	// 2. Ambil kunci simetris file asli
	fileKeyBytes, _ := hex.DecodeString(req.File.EncryptedKeyHex)

	// 3. ENKRIPSI KUNCI SIMETRIS DENGAN PUBLIC KEY PEMOHON (RSA-OAEP)
	encryptedKeyForReq, err := crypto.EncryptRSAOAEP(fileKeyBytes, requesterPubKeyPEM)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendistribusikan kunci: " + err.Error()})
		return
	}

	// 4. Simpan kunci terenkripsi ke database
	encKeyHex := hex.EncodeToString(encryptedKeyForReq)
	if err := h.vaultRepo.ApproveShareRequest(req.ID, encKeyHex); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyetujui permohonan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Akses disetujui! Kunci simetris telah diamankan dengan RSA Public Key penerima.",
	})
}

// 7. GetMyVaultFiles: Mengambil daftar file milik user yang sedang login
func (h *VaultHandler) GetMyVaultFiles(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	files, err := h.vaultRepo.GetFilesByOwner(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar berkas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files})
}

// RevokeShareAccess (Pemilik) mencabut hak akses berkas yang sebelumnya diberikan
func (h *VaultHandler) RevokeShareAccess(c *gin.Context) {
	ownerID := c.MustGet("user_id").(uint)
	reqIDStr := c.Param("id")
	reqID, _ := strconv.ParseUint(reqIDStr, 10, 32)

	req, err := h.vaultRepo.GetShareRequestByID(uint(reqID))
	if err != nil || req == nil || req.OwnerID != ownerID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data izin berbagi tidak ditemukan"})
		return
	}

	if err := h.vaultRepo.RevokeShareRequest(req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencabut izin akses"})
		return
	}

	// Catat di Activity Log
	h.vaultRepo.LogActivity(ownerID, "REVOKE", fmt.Sprintf("Mencabut akses berkas ID %d untuk user ID %d", req.FileID, req.RequesterID))

	c.JSON(http.StatusOK, gin.H{"message": "Hak akses berkas berhasil dicabut!"})
}