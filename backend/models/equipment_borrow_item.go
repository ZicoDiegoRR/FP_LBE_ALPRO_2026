package models

// EquipmentBorrowItem links a Borrow to the LabEquipment items it covers.
// There's no direct Lab_id on Borrow (schema is already frozen), so finding
// which lab a borrow belongs to means going: Borrow -> EquipmentBorrowItem
// -> LabEquipment -> Lab. All items on one borrow share the same lab, so
// looking at just the first item is enough.
type EquipmentBorrowItem struct {
	BorrowID       string `gorm:"column:borrow_id;primaryKey;size:20"`
	LabEquipmentID string `gorm:"column:lab_equipment_id;primaryKey;size:20"`
}

func (EquipmentBorrowItem) TableName() string { return "equipmentborrowitem" }
