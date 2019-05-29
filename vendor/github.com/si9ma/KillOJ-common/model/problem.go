package model

import (
	"time"
)

type Problem struct {
	ID               int               `gorm:"column:id;primary_key" json:"id"`
	Name             string            `gorm:"column:name" json:"name" binding:"required,max=100"`
	CreatedAt        time.Time         `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time         `gorm:"column:updated_at" json:"updated_at"`
	OwnerID          int               `gorm:"column:owner_id" json:"owner_id"`
	Desc             string            `gorm:"column:desc" json:"desc" binding:"required"`
	Input            string            `gorm:"column:input" json:"input" binding:"required"`
	Output           string            `gorm:"column:output" json:"output" binding:"required"`
	Hint             string            `gorm:"column:hint" json:"hint"`
	Source           string            `gorm:"column:source" json:"source" binding:"max=100"`
	TimeLimit        int               `gorm:"column:time_limit" json:"time_limit"`
	MemoryLimit      int               `gorm:"column:memory_limit" json:"memory_limit"`
	Difficulty       Difficulty        `gorm:"column:difficulty" json:"difficulty" binding:"exists,oneof=0 1 2"`
	BelongType       BelongType        `gorm:"column:belong_type" json:"belong_type" binding:"exists,oneof=0 1 2"`
	BelongToID       int               `gorm:"column:belong_to_id" json:"belong_to_id"`
	CatalogID        int               `gorm:"column:catalog_id" json:"catalog_id" binding:"required"`
	Catalog          Catalog           `json:"catalog" binding:"-"`
	Tags             []Tag             `gorm:"many2many:problem_has_tag;" json:"tags" binding:"dive"`
	ProblemSamples   []ProblemSample   `json:"samples" binding:"dive"`
	ProblemTestCases []ProblemTestCase `json:"test_cases,omitempty" binding:"required,dive"`
	UpVoteUsers      []UserWithOnlyID  `gorm:"many2many:user_vote_problem;association_jointable_foreignkey:user_id;" json:"up_vote_users"`
	DownVoteUsers    []UserWithOnlyID  `gorm:"many2many:user_vote_problem;association_jointable_foreignkey:user_id;" json:"down_vote_users"`
	Comments         []Comment         `json:"comments"`
	Owner            User              `json:"owner" gorm:"foreignkey:OwnerID;association_autoupdate:false;association_autocreate:false" binding:"-"`
	Limit            JSON              `gorm:"column:limit" json:"limit" binding:"required"` // todo store json in db may unreasonable
}

// TableName sets the insert table name for this struct type
func (p *Problem) TableName() string {
	return "problem"
}

type Limit struct {
	TimeLimit   int `json:"time_limit"`
	MemoryLimit int `json:"memory_limit"`
}

type BelongType int

const (
	BelongToPublic = BelongType(iota)
	BelongToGroup
	BelongToContest
)

type Difficulty int

const (
	Easy = Difficulty(iota)
	Medium
	Hard
)
