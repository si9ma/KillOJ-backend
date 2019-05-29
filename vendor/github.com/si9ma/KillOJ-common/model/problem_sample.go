package model

import (
	"encoding/json"
	"time"
)

type ProblemSample struct {
	ID        int       `gorm:"column:id;primary_key" json:"id" binding:"requiredwhenfield=DeleteIt"`
	ProblemID int       `gorm:"column:problem_id" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
	Input     string    `gorm:"column:input" json:"input" binding:"required"`
	Output    string    `gorm:"column:output" json:"output" binding:"required"`
	DeleteIt  bool      `gorm:"-" json:"delete_it,omitempty"`
}

// TableName sets the insert table name for this struct type
func (p *ProblemSample) TableName() string {
	return "problem_sample"
}

// ignore delete_it
func (p *ProblemSample) MarshalJSON() ([]byte, error) {
	type SampleAlias ProblemSample
	return json.Marshal(&struct {
		*SampleAlias
		DeleteIt bool `json:"delete_it,omitempty"`
	}{
		SampleAlias: (*SampleAlias)(p),
	})
}
