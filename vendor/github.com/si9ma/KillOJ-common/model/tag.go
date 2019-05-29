package model

import (
	"encoding/json"
	"time"
)

type Tag struct {
	ID        int       `gorm:"column:id;primary_key" json:"id" binding:"requiredwhenfield=DeleteIt"`
	Name      string    `gorm:"column:name" json:"name" binding:"required,min=1,max=50"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
	DeleteIt  bool      `gorm:"-" json:"delete_it,omitempty"`
}

// TableName sets the insert table name for this struct type
func (t *Tag) TableName() string {
	return "tag"
}

// ignore delete_it
func (t *Tag) MarshalJSON() ([]byte, error) {
	type TagAlias Tag
	return json.Marshal(&struct {
		*TagAlias
		DeleteIt bool `json:"delete_it,omitempty"`
	}{
		TagAlias: (*TagAlias)(t),
	})
}
