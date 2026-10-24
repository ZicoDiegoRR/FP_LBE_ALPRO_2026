package models

import "time"

type Borrow struct {
	ID                      string     `gorm:"column:id;primaryKey;size:20"`
	ScheduledDatetime       time.Time  `gorm:"column:scheduleddatetime"`
	BookedDatetime          time.Time  `gorm:"column:bookeddatetime"`
	Reason                  string     `gorm:"column:reason;size:100"`
	BorrowDurationSeconds   int        `gorm:"column:borrowdurationseconds"`
	BorrowStatusID          string     `gorm:"column:borrowstatus_id;size:20"`
	UserID                  string     `gorm:"column:user_id;size:20"`
	FinishedAt              *time.Time `gorm:"column:finishedat"`
	Title                   string     `gorm:"column:title;size:100"`
	StatusMessage           string     `gorm:"column:statusmessage;size:100"`
	StatusMessageUpdateTime time.Time  `gorm:"column:statusmessageupdatetime"`
}

func (Borrow) TableName() string { return "borrow" }