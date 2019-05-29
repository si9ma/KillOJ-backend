package model

import (
	"database/sql"
	"time"

	"github.com/guregu/null"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
)

type ProblemUpdateLog struct {
	LogID     int    `gorm:"column:log_id;primary_key" json:"log_id"`
	UserID    int    `gorm:"column:user_id" json:"user_id"`
	ProblemID int    `gorm:"column:problem_id" json:"problem_id"`
	BeforeLog string `gorm:"column:before_log" json:"before_log"`
	Commit    string `gorm:"column:commit" json:"commit"`
	AfterLog  string `gorm:"column:after_log" json:"after_log"`
}

// TableName sets the insert table name for this struct type
func (p *ProblemUpdateLog) TableName() string {
	return "problem_update_log"
}
