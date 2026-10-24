package models

type EquipmentStatus struct {
	ID     string `gorm:"column:id;primaryKey;size:20"`
	Name   string `gorm:"column:name;size:20"`
	Desc   string `gorm:"column:desc;size:100"`
	Symbol string `gorm:"column:symbol;size:1"`
}

func (EquipmentStatus) TableName() string { return "equipmentstatus" }
