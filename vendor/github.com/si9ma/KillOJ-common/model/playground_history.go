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

type PlaygroundHistory struct {
	ID           int         `gorm:"column:id;primary_key" json:"id"`
	SourceCode   null.String `gorm:"column:source_code" json:"source_code"`
	PlaygroundID int         `gorm:"column:playground_id" json:"playground_id"`
}

// TableName sets the insert table name for this struct type
func (p *PlaygroundHistory) TableName() string {
	return "playground_history"
}
