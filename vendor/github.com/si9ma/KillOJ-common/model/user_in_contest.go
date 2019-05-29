package model

import (
	"time"
)

type UserInContest struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	ContestID int       `gorm:"column:contest_id" json:"contest_id"`
	UserID    int       `gorm:"column:user_id" json:"user_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
}

// TableName sets the insert table name for this struct type
func (u *UserInContest) TableName() string {
	return "user_in_contest"
}
