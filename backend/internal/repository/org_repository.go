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

// --- 1. ORGANISASI & KEBIJAKAN FORM KUSTOM ---

func (r *OrgRepository) CreateOrganization(org *model.Organization) error {
	return r.db.Create(org).Error
}

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

func (r *OrgRepository) GetOrgByID(id uint) (*model.Organization, error) {
	var org model.Organization
	err := r.db.First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// UpdateOrgCustomSchema memperbarui skema form dinamis & kebijakan berkas
func (r *OrgRepository) UpdateOrgCustomSchema(orgID uint, schemaJSON string, requireDoc bool, docType string) error {
	return r.db.Model(&model.Organization{}).Where("id = ?", orgID).Updates(map[string]interface{}{
		"custom_fields_schema": schemaJSON,
		"require_document":     requireDoc,
		"document_type":        docType,
	}).Error
}

// --- 2. KEANGGOTAAN / DYNAMIC BIODATA ---

func (r *OrgRepository) RequestJoinOrg(member *model.OrganizationMember) error {
	return r.db.Create(member).Error
}

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

func (r *OrgRepository) GetPendingMembersByOrg(orgID uint) ([]model.OrganizationMember, error) {
	var members []model.OrganizationMember
	err := r.db.Preload("User").Where("organization_id = ?", orgID).Order("created_at desc").Find(&members).Error
	return members, err
}

func (r *OrgRepository) UpdateMemberStatus(memberID uint, status string) error {
	return r.db.Model(&model.OrganizationMember{}).Where("id = ?", memberID).Update("status", status).Error
}

// --- 3. INVENTARIS DENGAN LOGIKA DEDUPLIKASI STOK ---

// AddOrIncrementEquipment menambahkan alat baru, atau menambah kuantitas jika alat sudah terdaftar
func (r *OrgRepository) AddOrIncrementEquipment(orgID uint, code, name, category, desc string, qty int) (*model.Equipment, error) {
	var existing model.Equipment
	// Cari apakah alat dengan Kode atau Nama yang sama sudah ada di organisasi ini
	err := r.db.Where("organization_id = ? AND (code = ? OR name = ?)", orgID, code, name).First(&existing).Error

	if err == nil {
		// ALAT SUDAH ADA: Lakukan increment kuantitas (Deduplikasi)
		existing.TotalQuantity += qty
		existing.AvailableQuantity += qty
		if desc != "" {
			existing.Description = desc
		}
		if err := r.db.Save(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ALAT BELUM ADA: Buat entri baru
	newEq := model.Equipment{
		OrganizationID:    orgID,
		Code:              code,
		Name:              name,
		Category:          category,
		Description:       desc,
		TotalQuantity:     qty,
		AvailableQuantity: qty,
		Condition:         "BAIK",
		Status:            "AVAILABLE",
	}

	if err := r.db.Create(&newEq).Error; err != nil {
		return nil, err
	}
	return &newEq, nil
}

func (r *OrgRepository) GetEquipmentsByOrg(orgID uint) ([]model.Equipment, error) {
	var equipments []model.Equipment
	err := r.db.Where("organization_id = ?", orgID).Order("created_at desc").Find(&equipments).Error
	return equipments, err
}

func (r *OrgRepository) GetEquipmentByID(id uint) (*model.Equipment, error) {
	var eq model.Equipment
	err := r.db.First(&eq, id).Error
	if err != nil {
		return nil, err
	}
	return &eq, nil
}