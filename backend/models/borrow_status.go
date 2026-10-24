package models

type BorrowStatus struct {
	ID   string `gorm:"column:id;primaryKey;size:20"`
	Name string `gorm:"column:name;size:20"`
	Desc string `gorm:"column:desc;size:100"`
}

func (BorrowStatus) TableName() string { return "borrowstatus" }
