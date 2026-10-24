package models

type Equipment struct {
	ID         string `gorm:"column:id;primaryKey;size:20"`
	Name       string `gorm:"column:name;size:30"` // the general equipment "type", e.g. "Computer"
	Desc       string `gorm:"column:desc;size:100"`
	PictureURL string `gorm:"column:pictureurl;size:256"`
}

func (Equipment) TableName() string { return "equipment" }
