package models

type Role struct {
	ID            string `gorm:"column:id;primaryKey;size:20"`
	Name          string `gorm:"column:name;size:10"`
	IsSystemAdmin bool   `gorm:"column:issystemadmin"`
	IsLabAdmin    bool   `gorm:"column:islabadmin"`
}

func (Role) TableName() string {
	return "role"
}
