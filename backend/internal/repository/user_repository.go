package repository

import (
	"errors"

	"gorm.io/gorm"
	"pmo-backend/internal/model"
)

// UserRepository mendefinisikan kontrak operasi database untuk User
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository membuat instance baru UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser menyimpan user baru ke database
func (r *UserRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByUsername mencari data user berdasarkan username
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil jika user tidak ditemukan (bukan crash)
		}
		return nil, err
	}
	return &user, nil
}

// GetByID mencari data user berdasarkan Primary Key ID
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}