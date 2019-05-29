package model

import (
	"time"
)

type Contest struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	Name      string    `gorm:"column:name" json:"name" binding:"required,min=1,max=50"`
	OwnerID   int       `gorm:"column:owner_id" json:"owner_id"`
	StartTime time.Time `gorm:"column:start_time" json:"start_time" binding:"required,gte"`
	EndTime   time.Time `gorm:"column:end_time" json:"end_time" binding:"required,gtfield=StartTime"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
	Users     []User    `gorm:"many2many:user_in_contest;" json:"-"`
	Problems  []Problem `gorm:"foreignkey:BelongToID" json:"-"`
	Owner     User      `json:"owner" gorm:"foreignkey:OwnerID;association_autoupdate:false;association_autocreate:false" binding:"-"`
}

// TableName sets the insert table name for this struct type
func (c *Contest) TableName() string {
	return "contest"
}
