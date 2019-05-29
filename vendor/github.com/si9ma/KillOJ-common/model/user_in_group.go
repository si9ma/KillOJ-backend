package model

import (
	"time"
)

type UserInGroup struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	GroupID   int       `gorm:"column:group_id" json:"group_id"`
	UserID    int       `gorm:"column:user_id" json:"user_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName sets the insert table name for this struct type
func (u *UserInGroup) TableName() string {
	return "user_in_group"
}
