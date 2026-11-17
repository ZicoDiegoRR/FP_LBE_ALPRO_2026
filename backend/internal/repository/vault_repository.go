package repository

import (
	"errors"

	"gorm.io/gorm"
	"pmo-backend/internal/model"

	"time"
)

type VaultRepository struct {
	db *gorm.DB
}

func NewVaultRepository(db *gorm.DB) *VaultRepository {
	return &VaultRepository{db: db}
}

// --- 1. KEY MANAGEMENT (RSA KEY PAIR) ---

// SaveKeyPair menyimpan pasangan kunci RSA pengguna
func (r *VaultRepository) SaveKeyPair(keyPair *model.UserKeyPair) error {
	return r.db.Create(keyPair).Error
}

// GetKeyPairByUserID mengambil pasangan kunci berdasarkan User ID
func (r *VaultRepository) GetKeyPairByUserID(userID uint) (*model.UserKeyPair, error) {
	var keyPair model.UserKeyPair
	err := r.db.Where("user_id = ?", userID).First(&keyPair).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &keyPair, nil
}

// GetPublicKeyByUserID mengambil hanya Public Key milik user lain (untuk enkripsi kunci bersama)
func (r *VaultRepository) GetPublicKeyByUserID(userID uint) (string, error) {
	var keyPair model.UserKeyPair
	err := r.db.Select("public_key_pem").Where("user_id = ?", userID).First(&keyPair).Error
	if err != nil {
		return "", err
	}
	return keyPair.PublicKeyPEM, nil
}

// --- 2. VAULT FILES (BERKAS PRIBADI & TANDA TANGAN DIGITAL) ---

// SaveFile menyimpan metadata berkas terenkripsi ke database
func (r *VaultRepository) SaveFile(file *model.VaultFile) error {
	return r.db.Create(file).Error
}

// GetFilesByOwner mengambil daftar berkas brankas milik user yang login
func (r *VaultRepository) GetFilesByOwner(ownerID uint) ([]model.VaultFile, error) {
	var files []model.VaultFile
	err := r.db.Where("owner_id = ?", ownerID).Order("created_at desc").Find(&files).Error
	return files, err
}

// GetFileByID mengambil data file spesifik berdasarkan ID
func (r *VaultRepository) GetFileByID(fileID uint) (*model.VaultFile, error) {
	var file model.VaultFile
	err := r.db.Preload("Owner").First(&file, fileID).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// UpdateSignature mencatat status tanda tangan digital pada berkas
func (r *VaultRepository) UpdateSignature(fileID uint, signatureHex string) error {
	now := gorm.Expr("CURRENT_TIMESTAMP")
	return r.db.Model(&model.VaultFile{}).Where("id = ?", fileID).Updates(map[string]interface{}{
		"is_signed":     true,
		"signature_hex": signatureHex,
		"signed_at":     now,
	}).Error
}

// --- 3. SECURE FILE SHARING (PERMINTAAN & DISTRIBUSI KUNCI) ---

// CreateShareRequest membuat permintaan akses berkas baru (status: PENDING)
func (r *VaultRepository) CreateShareRequest(req *model.FileShareRequest) error {
	return r.db.Create(req).Error
}

// GetIncomingRequests mengambil daftar permintaan berkas yang masuk ke pemilik file (Owner)
func (r *VaultRepository) GetIncomingRequests(ownerID uint) ([]model.FileShareRequest, error) {
	var requests []model.FileShareRequest
	err := r.db.Preload("File").Preload("Requester").Where("owner_id = ?", ownerID).Order("created_at desc").Find(&requests).Error
	return requests, err
}

// GetOutgoingRequests mengambil daftar permintaan berkas yang diajukan oleh pemohon (Requester)
func (r *VaultRepository) GetOutgoingRequests(requesterID uint) ([]model.FileShareRequest, error) {
	var requests []model.FileShareRequest
	err := r.db.Preload("File").Preload("Owner").Where("requester_id = ?", requesterID).Order("created_at desc").Find(&requests).Error
	return requests, err
}

// GetShareRequestByID mengambil permohonan spesifik berdasarkan ID
func (r *VaultRepository) GetShareRequestByID(requestID uint) (*model.FileShareRequest, error) {
	var req model.FileShareRequest
	err := r.db.Preload("File").Preload("Requester").Preload("Owner").First(&req, requestID).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// ApproveShareRequest menyetujui akses dan menyimpan kunci simetris terenkripsi (RSA) untuk pemohon
func (r *VaultRepository) ApproveShareRequest(requestID uint, encryptedKeyForReq string) error {
	return r.db.Model(&model.FileShareRequest{}).Where("id = ?", requestID).Updates(map[string]interface{}{
		"status":                 "APPROVED",
		"encrypted_key_for_req": encryptedKeyForReq,
	}).Error
}

// RejectShareRequest menolak permintaan akses berkas
func (r *VaultRepository) RejectShareRequest(requestID uint) error {
	return r.db.Model(&model.FileShareRequest{}).Where("id = ?", requestID).Update("status", "REJECTED").Error
}

// LogActivity mencatat jejak audit aktivitas berkas/kriptografi
func (r *VaultRepository) LogActivity(userID uint, action, details string) error {
	log := model.ActivityLog{
		UserID:    userID,
		Action:    action,
		Details:   details,
		CreatedAt: time.Now(),
	}
	return r.db.Create(&log).Error
}

// RevokeShareRequest mencabut hak akses berkas yang sebelumnya telah disetujui
func (r *VaultRepository) RevokeShareRequest(requestID uint) error {
	now := time.Now()
	return r.db.Model(&model.FileShareRequest{}).Where("id = ?", requestID).Updates(map[string]interface{}{
		"status":     "REVOKED",
		"revoked_at": &now,
	}).Error
}
