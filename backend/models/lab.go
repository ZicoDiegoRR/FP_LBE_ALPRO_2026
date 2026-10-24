package models

type Lab struct {
	ID         string `gorm:"column:id;primaryKey;size:20"`
	Name       string `gorm:"column:name;size:30"`
	Desc       string `gorm:"column:desc;size:100"`
	PictureURL string `gorm:"column:pictureurl;size:256"`
}

func (Lab) TableName() string { return "lab" }
