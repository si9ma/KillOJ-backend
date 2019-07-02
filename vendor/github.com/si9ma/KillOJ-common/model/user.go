package model

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           int       `gorm:"column:id;primary_key" json:"id"`
	GithubUserID string    `gorm:"column:github_user_id" json:"github_user_id,omitempty"`
	GithubName   string    `gorm:"column:github_name" json:"github_name"`
	Email        string    `gorm:"column:email" json:"email" binding:"required,email,max=100"`
	Name         string    `gorm:"column:name" json:"name" binding:"required,max=100,excludesall=!@#?"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"-"`
	Role         int       `gorm:"column:role" json:"role"`

	// password from user input, password should't response to user,
	// so set omitempty && set this field to nil before return
	Password         string `gorm:"-" json:"password,omitempty" binding:"omitempty,min=6,max=30"`
	EncryptedPasswd  string `gorm:"column:passwd" json:"-"` // encrypted password in db
	Signature        string `gorm:"column:signature" json:"signature" binding:"max=800"`
	NoInOrganization string `gorm:"column:no_in_organization" json:"no_in_organization" binding:"max=30"`
	Organization     string `gorm:"column:organization" json:"organization" binding:"max=50"`
	Site             string `gorm:"column:site" json:"site" binding:"max=50"`
	AvatarUrl        string `gorm:"column:avatar_url" json:"avatar" binding:"max=600"`
	WeiboName        string `gorm:"column:weibo_name" json:"weibo_name" binding:"max=50"`
	ZhihuName        string `gorm:"column:zhihu_name" json:"zhihu_name" binding:"max=50"`

	// theme
	Theme *Theme `json:"theme,omitempty" binding:"-"`

	// group
	Groups   []Group   `gorm:"many2many:user_in_group;" json:"-"`
	Contests []Contest `gorm:"many2many:user_in_contest;" json:"-"`
}

// TableName sets the insert table name for this struct type
func (u *User) TableName() string {
	return "user"
}

// ignore password and github user id
func (u *User) MarshalJSON() ([]byte, error) {
	type UserAlias User
	return json.Marshal(&struct {
		*UserAlias
		Password     string `json:"password,omitempty"`
		GithubUserID string `json:"github_user_id,omitempty"`
	}{
		UserAlias: (*UserAlias)(u),
	})
}

type Role int

const (
	Administrator = Role(iota)
	Maintainer
	Normal
)

type UserWithOnlyID struct {
	ID int `gorm:"column:id;primary_key" json:"id"`
}

// TableName sets the insert table name for this struct type
func (u *UserWithOnlyID) TableName() string {
	return "user"
}
