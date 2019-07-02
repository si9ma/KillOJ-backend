package model

import (
	"time"
)

type Theme struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	UserID    int       `gorm:"column:user_id" json:"user_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
	Theme     int       `gorm:"column:theme" json:"theme" binding:"exists,oneof=0 1"`
	SidebarBg string    `gorm:"column:sidebar_bg" json:"sidebar_bg" binding:"required,oneof=vue green blue purple"`
	Direction int       `gorm:"column:direction" json:"direction" binding:"exists,oneof=0 1"`
}

// TableName sets the insert table name for this struct type
func (t *Theme) TableName() string {
	return "theme"
}
