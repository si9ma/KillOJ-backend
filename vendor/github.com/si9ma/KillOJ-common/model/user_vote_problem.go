package model

import (
	"time"
)

type UserVoteProblem struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
	UserID    int       `gorm:"column:user_id;primary_key" json:"-"`
	ProblemID int       `gorm:"column:problem_id;primary_key" json:"problem_id"`
	Attitude  int       `gorm:"column:attitude" json:"attitude" form:"attitude" binding:"oneof=-1 0 1"`
}

// TableName sets the insert table name for this struct type
func (u *UserVoteProblem) TableName() string {
	return "user_vote_problem"
}

type Attitude int

const (
	Down       = Attitude(-1)
	NoAttitude = Attitude(0)
	Up         = Attitude(1)
)
