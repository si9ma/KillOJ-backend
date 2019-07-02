package model

import "time"

type Group struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	OwnerID   int       `gorm:"column:owner_id" json:"owner_id"`
	Name      string    `gorm:"column:name" json:"name" binding:"required,max=50"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
	Users     []User    `gorm:"many2many:user_in_group;" json:"-"`
	Problems  []Problem `gorm:"foreignkey:BelongToID" json:"-"`
	Owner     User      `json:"owner" binding:"-"`
}

// TableName sets the insert table name for this struct type
func (g *Group) TableName() string {
	return "group"
}
