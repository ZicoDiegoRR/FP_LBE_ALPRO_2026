package model

import (
	"time"
)

// User merepresentasikan entitas pengguna di sistem
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Password  string    `gorm:"not null" json:"-"` // Tanda json:"-" mencegah hash password bocor ke respons API
	Role      string    `gorm:"default:'user';size:20" json:"role"` // 'user' atau 'admin'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}