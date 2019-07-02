package model

import (
	"time"
)

type Comment struct {
	ID         int       `gorm:"column:id;primary_key" json:"id"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
	FromID     int       `gorm:"column:from_id" json:"from_id"`
	ToID       int       `gorm:"column:to_id" json:"to_id"`
	ProblemID  int       `gorm:"column:problem_id" json:"problem_id"`
	ForComment int       `gorm:"column:for_comment" json:"for_comment"`
	Content    string    `gorm:"column:content" json:"content"`
	From       User      `json:"from" gorm:"foreignkey:FromID;" binding:"-"`
	To         User      `json:"to" gorm:"foreignkey:ToID;" binding:"-"`
}

// TableName sets the insert table name for this struct type
func (c *Comment) TableName() string {
	return "comment"
}
