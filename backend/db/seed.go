package db

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"pmo-backend/models"
)

// insertIgnore memasukkan baris hanya jika ID-nya belum ada,
// jadi aman dijalankan setiap kali server start.
func insertIgnore(d *gorm.DB, rows interface{}) {
	if err := d.Clauses(clause.OnConflict{DoNothing: true}).Create(rows).Error; err != nil {
		log.Fatalf("seed failed: %v", err)
	}
}

// ID di bawah HARUS sama dengan konstanta di handlers/dashboard.go,
// handlers/status_sync.go, dan handlers/auth.go.
func seed(d *gorm.DB) {
	insertIgnore(d, &[]models.Role{
		{ID: "F4L9SOWPVI7465NEOBM7", Name: "Member"},
		{ID: "ADMINROLE0000000001", Name: "Admin", IsSystemAdmin: true, IsLabAdmin: true},
	})

	insertIgnore(d, &[]models.BorrowStatus{
		{ID: "8371HA2RZB8EKYO7FU6T", Name: "Pending", Desc: "Waiting for approval"},
		{ID: "QTZ9IYVBZTAMKEIAEUZL", Name: "Accepted", Desc: "Approved by admin"},
		{ID: "W0ABU8J8NTBB7ERDLOF5", Name: "Rejected", Desc: "Rejected by admin"},
		{ID: "ITNNA5NGDAJ1JMBJ0VZJ", Name: "Cancelled", Desc: "Cancelled by user"},
		{ID: "AC0TBLTYQUYCSJ096T7R", Name: "Failed", Desc: "Not approved in time"},
		{ID: "0KHQJHJSO3DUHGU53IY5", Name: "Ongoing", Desc: "Currently borrowed"},
		{ID: "IKXUHNQBZG523YCGUB7J", Name: "Finished", Desc: "Returned"},
	})

	insertIgnore(d, &[]models.EquipmentStatus{
		{ID: "JK7QYOKLJO6CEOFG4UVV", Name: "Available", Desc: "Can be borrowed", Symbol: "✓"},
		{ID: "6OEF7PYAFQJZALEETSAW", Name: "In Use", Desc: "Currently booked", Symbol: "✗"},
	})

	insertIgnore(d, &[]models.Lab{
		{ID: "LAB-EMBEDDED", Name: "Embedded Systems Lab", Desc: "Microcontrollers and sensors"},
		{ID: "LAB-ELECTRONICS", Name: "Electronics Lab", Desc: "Measurement instruments"},
	})

	pic := "/equipment_pictures/computer.jpg"
	insertIgnore(d, &[]models.Equipment{
		{ID: "EQT-MCU", Name: "Microcontroller", Desc: "Dev boards", PictureURL: pic},
		{ID: "EQT-SENSOR", Name: "Sensor", Desc: "Sensors", PictureURL: pic},
		{ID: "EQT-SCOPE", Name: "Oscilloscope", Desc: "Signal measurement", PictureURL: pic},
		{ID: "EQT-METER", Name: "Multimeter", Desc: "Voltage and current", PictureURL: pic},
	})

	avail := "JK7QYOKLJO6CEOFG4UVV"
	insertIgnore(d, &[]models.LabEquipment{
		{ID: "LE-EMB-1", LabID: "LAB-EMBEDDED", EquipmentID: "EQT-MCU", Name: "ESP32 DevKit #1", Desc: "ESP32 board", EquipmentStatusID: avail},
		{ID: "LE-EMB-2", LabID: "LAB-EMBEDDED", EquipmentID: "EQT-MCU", Name: "ARM Cortex Board #1", Desc: "Cortex-M board", EquipmentStatusID: avail},
		{ID: "LE-EMB-3", LabID: "LAB-EMBEDDED", EquipmentID: "EQT-SENSOR", Name: "Temp Sensor #1", Desc: "Temperature sensor", EquipmentStatusID: avail},
		{ID: "LE-ELC-1", LabID: "LAB-ELECTRONICS", EquipmentID: "EQT-SCOPE", Name: "Oscilloscope #1", Desc: "2-channel scope", EquipmentStatusID: avail},
		{ID: "LE-ELC-2", LabID: "LAB-ELECTRONICS", EquipmentID: "EQT-METER", Name: "Multimeter #1", Desc: "Digital multimeter", EquipmentStatusID: avail},
	})
}