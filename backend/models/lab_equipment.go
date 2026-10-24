package models

type LabEquipment struct {
	ID                string `gorm:"column:id;primaryKey;size:20"`
	LabID             string `gorm:"column:lab_id;size:20"`
	EquipmentID       string `gorm:"column:equipment_id;size:20"`
	Name              string `gorm:"column:name;size:30"`
	Desc              string `gorm:"column:desc;size:100"`
	EquipmentStatusID string `gorm:"column:equipmentstatus_id;size:20"`
}

func (LabEquipment) TableName() string { return "labequipment" }
