package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/model"
	"pmo-backend/internal/repository"
)

type OrgHandler struct {
	orgRepo *repository.OrgRepository
}

func NewOrgHandler(orgRepo *repository.OrgRepository) *OrgHandler {
	return &OrgHandler{orgRepo: orgRepo}
}

// generateRedeemCode membuat 8 karakter acak alfanumerik unik untuk organisasi
func generateRedeemCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

// CreateOrganization (Khusus Admin) membuat organisasi baru
func (h *OrgHandler) CreateOrganization(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)

	// Cek apakah admin sudah memiliki organisasi
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

	newOrg := model.Organization{
		Name:        input.Name,
		RedeemCode:  generateRedeemCode(),
		OwnerID:     adminID,
		Description: input.Description,
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

// GetMyOrganization (Khusus Admin) mengambil detail organisasi miliknya
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

// JoinOrganization (Peminjam / User) mendaftar ke organisasi via Redeem Code + Biodata
func (h *OrgHandler) JoinOrganization(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Cek apakah user sudah terdaftar di suatu organisasi
	existingMember, _ := h.orgRepo.GetMembershipByUser(userID)
	if existingMember != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Anda sudah terdaftar atau sedang dalam proses verifikasi di organisasi"})
		return
	}

	var input struct {
		RedeemCode     string `json:"redeem_code" binding:"required"`
		Fullname       string `json:"fullname" binding:"required,min=3"`
		IdentityNumber string `json:"identity_number" binding:"required"`
		PhoneNumber    string `json:"phone_number" binding:"required"`
		Reason         string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pastikan semua biodata dan kode unik terisi dengan benar"})
		return
	}

	// Cari organisasi berdasarkan Redeem Code
	org, err := h.orgRepo.GetOrgByRedeemCode(strings.TrimSpace(input.RedeemCode))
	if err != nil || org == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kode redeem organisasi tidak valid"})
		return
	}

	membership := model.OrganizationMember{
		UserID:         userID,
		OrganizationID: org.ID,
		Fullname:       input.Fullname,
		IdentityNumber: input.IdentityNumber,
		PhoneNumber:    input.PhoneNumber,
		Reason:         input.Reason,
		Status:         "PENDING",
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

// GetMyMembership (Peminjam / User) mengecek status keanggotaan saat ini
func (h *OrgHandler) GetMyMembership(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	member, err := h.orgRepo.GetMembershipByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa keanggotaan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"membership": member})
}

// GetPendingMembers (Khusus Admin) melihat daftar calon peminjam yang mendaftar
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

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// VerifyMember (Khusus Admin) menyetujui (APPROVED) atau menolak (REJECTED) anggota
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

// AddEquipment (Khusus Admin) menambahkan barang ke inventaris organisasi
func (h *OrgHandler) AddEquipment(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	org, err := h.orgRepo.GetOrgByOwnerID(adminID)
	if err != nil || org == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya admin organisasi yang berhak menambah alat"})
		return
	}

	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama alat/barang wajib diisi"})
		return
	}

	eq := model.Equipment{
		OrganizationID: org.ID,
		Name:           input.Name,
		Description:    input.Description,
		Status:         "AVAILABLE",
	}

	if err := h.orgRepo.CreateEquipment(&eq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendaftarkan alat"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Barang berhasil ditambahkan ke inventaris!",
		"equipment": eq,
	})
}

// GetEquipments melihat daftar barang yang tersedia di organisasi
func (h *OrgHandler) GetEquipments(c *gin.Context) {
	orgIDStr := c.Query("org_id")
	var orgID uint

	if orgIDStr != "" {
		parsed, _ := strconv.ParseUint(orgIDStr, 10, 32)
		orgID = uint(parsed)
	} else {
		// Jika tidak ada query org_id, cari organisasi milik user saat ini
		userID := c.MustGet("user_id").(uint)
		role := c.MustGet("role").(string)

		if role == "admin" {
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