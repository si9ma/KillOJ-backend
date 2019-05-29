package model

import (
	"time"
)

type Template struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	Template  string    `gorm:"column:template" json:"template" binding:"required"`
	Language  int       `gorm:"column:language" json:"language" binding:"required,oneof=0 1 2 3"`
}

// TableName sets the insert table name for this struct type
func (t *Template) TableName() string {
	return "template"
}
