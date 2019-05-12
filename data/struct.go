// common data struct
package data

import "github.com/si9ma/KillOJ-common/model"

type InviteData struct {
	ID       string `json:"id"`
	GroupID  int    `json:"group_id"`
	Password string `json:"password,omitempty" binding:"max=30"`
	Timeout  int    `json:"timeout" binding:"required,min=3600,max=2592000"` // second , max = 30 day,min = a hour
}

type GroupWrap struct {
	model.Group
	NeedPassword bool   `json:"need_password"`
	Password     string `json:"-"`
}
