package model

import (
	"time"
)

type ProblemHasTag struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	TagID     int       `gorm:"column:tag_id" json:"tag_id"`
	ProblemID int       `gorm:"column:problem_id" json:"problem_id"`
}

// TableName sets the insert table name for this struct type
func (p *ProblemHasTag) TableName() string {
	return "problem_has_tag"
}
