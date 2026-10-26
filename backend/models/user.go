package models

import "time"

// User maps to the "User" table (quoted/capitalized in the schema since it's
// a reserved word in Postgres). Column names below are unquoted in the
// original SQL, so Postgres folds them to lowercase automatically.
type User struct {
	ID           string    `gorm:"column:id;primaryKey;size:20"`
	Name         string    `gorm:"column:name;size:30"`
	CreationDate time.Time `gorm:"column:creationdate"`
	Email        string    `gorm:"column:email;size:50"`
	Password     string    `gorm:"column:password;size:256"`
	PfpURL       string    `gorm:"column:pfpurl;size:256"`
	RoleID       string    `gorm:"column:role_id;size:20"`
}

// func (User) TableName() string {
// 	return `"User"` // keep quoted: capitalized + reserved word in Postgres
// }

func (User) TableName() string {
	return "users"
}
