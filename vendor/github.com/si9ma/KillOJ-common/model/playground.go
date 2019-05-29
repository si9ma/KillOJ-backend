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

type Playground struct {
	ID         int    `gorm:"column:id;primary_key" json:"id"`
	UserID     int    `gorm:"column:user_id" json:"user_id"`
	SourceCode string `gorm:"column:source_code" json:"source_code"`
	CodeName   string `gorm:"column:code_name" json:"code_name"`
	Language   int    `gorm:"column:language" json:"language"`
}

// TableName sets the insert table name for this struct type
func (p *Playground) TableName() string {
	return "playground"
}
