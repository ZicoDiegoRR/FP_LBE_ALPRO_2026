package repository

import (
	"errors"

	"gorm.io/gorm"
	"pmo-backend/internal/model"
)

type OrgRepository struct {
	db *gorm.DB
}

func NewOrgRepository(db *gorm.DB) *OrgRepository {
	return &OrgRepository{db: db}
}

// --- ORGANISASI (ADMIN) ---

// CreateOrganization membuat organisasi baru beserta redeem code unik
func (r *OrgRepository) CreateOrganization(org *model.Organization) error {
	return r.db.Create(org).Error
}

// GetOrgByOwnerID mengambil organisasi milik Admin tertentu
func (r *OrgRepository) GetOrgByOwnerID(ownerID uint) (*model.Organization, error) {
	var org model.Organization
	err := r.db.Where("owner_id = ?", ownerID).First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// GetOrgByRedeemCode mencari organisasi berdasarkan Redeem Code
func (r *OrgRepository) GetOrgByRedeemCode(code string) (*model.Organization, error) {
	var org model.Organization
	err := r.db.Where("redeem_code = ?", code).First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// --- KEANGGOTAAN / MEMBERSHIP (2 LAPIS VERIFIKASI) ---

// RequestJoinOrg mengajukan pendaftaran anggota baru (status: PENDING)
func (r *OrgRepository) RequestJoinOrg(member *model.OrganizationMember) error {
	return r.db.Create(member).Error
}

// GetMembershipByUser mengambil status keanggotaan aktif dari seorang user
func (r *OrgRepository) GetMembershipByUser(userID uint) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	err := r.db.Preload("Organization").Where("user_id = ?", userID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// GetPendingMembersByOrg mengambil daftar calon peminjam yang menunggu persetujuan Admin
func (r *OrgRepository) GetPendingMembersByOrg(orgID uint) ([]model.OrganizationMember, error) {
	var members []model.OrganizationMember
	err := r.db.Preload("User").Where("organization_id = ?", orgID).Order("created_at desc").Find(&members).Error
	return members, err
}

// UpdateMemberStatus menyetujui (APPROVED) atau menolak (REJECTED) keanggotaan
func (r *OrgRepository) UpdateMemberStatus(memberID uint, status string) error {
	return r.db.Model(&model.OrganizationMember{}).Where("id = ?", memberID).Update("status", status).Error
}

// --- INVENTARIS ALAT / EQUIPMENT ---

// CreateEquipment menambahkan barang baru milik organisasi
func (r *OrgRepository) CreateEquipment(eq *model.Equipment) error {
	return r.db.Create(eq).Error
}

// GetEquipmentsByOrg mengambil seluruh alat milik organisasi
func (r *OrgRepository) GetEquipmentsByOrg(orgID uint) ([]model.Equipment, error) {
	var equipments []model.Equipment
	err := r.db.Where("organization_id = ?", orgID).Order("created_at desc").Find(&equipments).Error
	return equipments, err
}