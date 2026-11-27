package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/config"
	"pmo-backend/internal/crypto"
	"pmo-backend/internal/model"
	"pmo-backend/internal/repository"
)

type OrgHandler struct {
	orgRepo       *repository.OrgRepository
	cryptoService *crypto.CryptoService
	cfg           *config.Config
}

func NewOrgHandler(orgRepo *repository.OrgRepository, cryptoService *crypto.CryptoService, cfg *config.Config) *OrgHandler {
	return &OrgHandler{
		orgRepo:       orgRepo,
		cryptoService: cryptoService,
		cfg:           cfg,
	}
}

func (h *OrgHandler) encryptField(plaintext string) (string, error) {
	if plaintext == "" || h.cryptoService == nil || h.cfg == nil {
		return plaintext, nil
	}
	return h.cryptoService.EncryptString(plaintext, "AES", []byte(strings.TrimSpace(h.cfg.CryptoKeyAES)))
}

func (h *OrgHandler) decryptField(stored string) string {
	if stored == "" || h.cryptoService == nil || h.cfg == nil {
		return stored
	}
	plain, err := h.cryptoService.DecryptString(stored, "AES", []byte(strings.TrimSpace(h.cfg.CryptoKeyAES)))
	if err != nil {
		// Fallback untuk data plaintext lama di database
		return stored
	}
	return plain
}

func (h *OrgHandler) decryptMember(member *model.OrganizationMember) {
	if member == nil {
		return
	}
	member.Fullname = h.decryptField(member.Fullname)
	member.CustomBiodataJSON = h.decryptField(member.CustomBiodataJSON)
}

func generateRedeemCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

// CreateOrganization membuat organisasi baru
func (h *OrgHandler) CreateOrganization(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)

	existingOrg, _ := h.orgRepo.GetOrgByOwnerID(adminID)
	if existingOrg != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Anda sudah memiliki organisasi yang aktif"})
		return
	}

	var input struct {
		Name        string `json:"name" binding:"required,min=3,max=100"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama organisasi minimal 3 karakter"})
		return
	}

	defaultSchema := `[{"key":"no_identitas","label":"Nomor Identitas (NIM/NIK)","type":"text","required":true},{"key":"no_hp","label":"No WhatsApp / HP","type":"text","required":true}]`

	newOrg := model.Organization{
		Name:               input.Name,
		RedeemCode:         generateRedeemCode(),
		OwnerID:            adminID,
		Description:        input.Description,
		CustomFieldsSchema: defaultSchema,
		RequireDocument:    false,
		DocumentType:       "Surat Izin / KTM",
	}

	if err := h.orgRepo.CreateOrganization(&newOrg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat organisasi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Organisasi berhasil dibuat!",
		"organization": newOrg,
	})
}

// GetMyOrganization mengambil profil organisasi milik Admin
func (h *OrgHandler) GetMyOrganization(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	org, err := h.orgRepo.GetOrgByOwnerID(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data organisasi"})
		return
	}
	if org == nil {
		c.JSON(http.StatusOK, gin.H{"organization": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"organization": org})
}

// UpdateCustomSchema (Admin) mengatur form kustom (Privacy by Design) & syarat dokumen
func (h *OrgHandler) UpdateCustomSchema(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	org, err := h.orgRepo.GetOrgByOwnerID(adminID)
	if err != nil || org == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organisasi tidak ditemukan"})
		return
	}

	var input struct {
		CustomFieldsSchema string `json:"custom_fields_schema" binding:"required"`
		RequireDocument    bool   `json:"require_document"`
		DocumentType       string `json:"document_type"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data skema form tidak valid"})
		return
	}

	if err := h.orgRepo.UpdateOrgCustomSchema(org.ID, input.CustomFieldsSchema, input.RequireDocument, input.DocumentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui konfigurasi organisasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skema formulir & kebijakan berkas berhasil diperbarui!"})
}

// JoinOrganization (Peminjam) mendaftar dengan biodata dinamis
func (h *OrgHandler) JoinOrganization(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	existingMember, _ := h.orgRepo.GetMembershipByUser(userID)
	if existingMember != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Anda sudah terdaftar atau sedang dalam proses verifikasi"})
		return
	}

	var input struct {
		RedeemCode        string `json:"redeem_code" binding:"required"`
		Fullname          string `json:"fullname" binding:"required,min=3"`
		CustomBiodataJSON string `json:"custom_biodata_json"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode redeem dan Nama Lengkap wajib diisi"})
		return
	}

	org, err := h.orgRepo.GetOrgByRedeemCode(strings.TrimSpace(input.RedeemCode))
	if err != nil || org == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kode redeem organisasi tidak valid"})
		return
	}

	encFullname, err := h.encryptField(input.Fullname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi nama lengkap"})
		return
	}
	encBiodata, err := h.encryptField(input.CustomBiodataJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi biodata anggota"})
		return
	}

	membership := model.OrganizationMember{
		UserID:            userID,
		OrganizationID:    org.ID,
		Fullname:          encFullname,
		CustomBiodataJSON: encBiodata,
		Status:            "PENDING",
	}

	if err := h.orgRepo.RequestJoinOrg(&membership); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengajukan keanggotaan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengajuan bergabung berhasil dikirim! Menunggu verifikasi persetujuan Admin.",
		"status":  "PENDING",
	})
}

// GetMyMembership melihat status keanggotaan peminjam
func (h *OrgHandler) GetMyMembership(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	member, err := h.orgRepo.GetMembershipByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa keanggotaan"})
		return
	}
	h.decryptMember(member)
	c.JSON(http.StatusOK, gin.H{"membership": member})
}

// GetPendingMembers (Admin) melihat daftar calon peminjam
func (h *OrgHandler) GetPendingMembers(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	org, err := h.orgRepo.GetOrgByOwnerID(adminID)
	if err != nil || org == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organisasi tidak ditemukan"})
		return
	}

	members, err := h.orgRepo.GetPendingMembersByOrg(org.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar anggota"})
		return
	}

	for i := range members {
		h.decryptMember(&members[i])
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// VerifyMember (Admin) menyetujui atau menolak anggota
func (h *OrgHandler) VerifyMember(c *gin.Context) {
	memberIDStr := c.Param("id")
	memberID, _ := strconv.ParseUint(memberIDStr, 10, 32)

	var input struct {
		Status string `json:"status" binding:"required,oneof=APPROVED REJECTED"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status harus APPROVED atau REJECTED"})
		return
	}

	if err := h.orgRepo.UpdateMemberStatus(uint(memberID), input.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status anggota"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status verifikasi anggota berhasil diperbarui!"})
}

// AddEquipment (Admin) menambah alat dengan Kuantitas & Deduplikasi Otomatis
func (h *OrgHandler) AddEquipment(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	org, err := h.orgRepo.GetOrgByOwnerID(adminID)
	if err != nil || org == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya admin organisasi yang berhak menambah alat"})
		return
	}

	var input struct {
		Code        string `json:"code" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Quantity    int    `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode, Nama barang, dan Kuantitas (minimal 1) wajib diisi"})
		return
	}

	category := input.Category
	if category == "" {
		category = "Umum"
	}

	eq, err := h.orgRepo.AddOrIncrementEquipment(org.ID, input.Code, input.Name, category, input.Description, input.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan inventaris: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Inventaris barang berhasil diperbarui!",
		"equipment": eq,
	})
}

// GetEquipments melihat daftar barang inventaris
func (h *OrgHandler) GetEquipments(c *gin.Context) {
	orgIDStr := c.Query("org_id")
	var orgID uint

	if orgIDStr != "" {
		parsed, _ := strconv.ParseUint(orgIDStr, 10, 32)
		orgID = uint(parsed)
	} else {
		userID := c.MustGet("user_id").(uint)
		role := c.MustGet("role").(string)

		if strings.EqualFold(role, "admin") {
			org, _ := h.orgRepo.GetOrgByOwnerID(userID)
			if org != nil {
				orgID = org.ID
			}
		} else {
			member, _ := h.orgRepo.GetMembershipByUser(userID)
			if member != nil && member.Status == "APPROVED" {
				orgID = member.OrganizationID
			}
		}
	}

	if orgID == 0 {
		c.JSON(http.StatusOK, gin.H{"equipments": []model.Equipment{}})
		return
	}

	equipments, err := h.orgRepo.GetEquipmentsByOrg(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar alat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"equipments": equipments})
}