package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/data"

	"github.com/si9ma/KillOJ-backend/srv"

	"github.com/si9ma/KillOJ-backend/auth"
	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/gin-gonic/gin"

	"github.com/si9ma/KillOJ-common/model"
)

func setupGroup(r *gin.Engine) {
	// need auth
	auth.AuthGroup.GET("/groups", GetAllGroups)
	auth.AuthGroup.GET("/groups/group/:id", GetGroup)
	auth.AuthGroup.POST("/groups", AddGroup)
	auth.AuthGroup.PUT("/groups/group/:id", UpdateGroup)
	auth.AuthGroup.POST("/groups/group/:id/invite", Invite)
	auth.AuthGroup.GET("/groups/join/:uuid", JoinQuery)
	auth.AuthGroup.POST("/groups/join/:uuid", Join)
	//auth.AuthGroup.DELETE("/groups/:id", DeleteGroup)
}

func GetAllGroups(c *gin.Context) {
	var err error
	ctx := c.Request.Context()
	arg := PageArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}

	groups, err := srv.GetAllGroups(c, arg.Page, arg.PageSize, arg.Order)
	if err != nil {
		log.For(ctx).Error("get groups fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, groups)
}

func GetGroup(c *gin.Context) {
	ctx := c.Request.Context()
	arg := QueryArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, true) {
		return
	}

	group, err := srv.GetGroup(c, arg.ID)
	if err != nil {
		log.For(ctx).Error("get group info fail", zap.Error(err), zap.Int("groupId", arg.ID))
		return
	}

	c.JSON(http.StatusOK, group)
}

func AddGroup(c *gin.Context) {
	ctx := c.Request.Context()
	newGroup := model.Group{}
	userId := auth.GetUserFromJWT(c).ID

	// bind
	if !wrap.ShouldBind(c, &newGroup, false) {
		return
	}

	newGroup.OwnerID = userId
	if err := srv.AddGroup(c, &newGroup); err != nil {
		log.For(ctx).Error("add group fail", zap.Error(err), zap.Int("groupId", newGroup.ID))
		return
	}

	c.JSON(http.StatusOK, newGroup)
}

func UpdateGroup(c *gin.Context) {
	ctx := c.Request.Context()
	newGroup := model.Group{}
	uriArg := QueryArg{}

	// bind uri params
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind request params
	if !wrap.ShouldBind(c, &newGroup, false) {
		return
	}

	// use id in uri path
	newGroup.ID = uriArg.ID
	if err := srv.UpdateGroup(c, &newGroup); err != nil {
		log.For(ctx).Error("update group fail", zap.Error(err), zap.Int("groupId", newGroup.ID))
		return
	}

	c.JSON(http.StatusOK, newGroup)
}

//
//func DeleteGroup(c *gin.Context) {
//	ctx := c.Request.Context()
//	uriArg := QueryArg{}
//
//	// bind uri params
//	if !wrap.ShouldBind(c, &uriArg, true) {
//		return
//	}
//
//	if err := srv.DeleteGroup(c, uriArg.ID); err != nil {
//		log.For(ctx).Error("delete group fail", zap.Error(err), zap.Int("groupId", uriArg.ID))
//		return
//	}
//
//	c.JSON(http.StatusOK, nil)
//}

func Invite(c *gin.Context) {
	ctx := c.Request.Context()
	inviteData := data.InviteData{}
	uriArg := QueryArg{}

	// bind uri
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind
	if !wrap.ShouldBind(c, &inviteData, false) {
		return
	}

	inviteData.GroupID = uriArg.ID
	if err := srv.Invite(c, &inviteData); err != nil {
		log.For(ctx).Error("group invite fail", zap.Error(err), zap.Int("groupId", uriArg.ID))
		return
	}

	// clear password before return
	inviteData.Password = ""
	c.JSON(http.StatusOK, inviteData)
}

type uuidArg struct {
	UUID string `uri:"uuid" binding:"uuid,required"`
}

func JoinQuery(c *gin.Context) {
	ctx := c.Request.Context()
	uriArg := uuidArg{}

	// bind uri
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	resp, err := srv.JoinQuery(c, uriArg.UUID)
	if err != nil {
		log.For(ctx).Error("query before join fail", zap.Error(err), zap.String("uuid", uriArg.UUID))
		return
	}

	c.JSON(http.StatusOK, resp)
}

type joinArg struct {
	Password string `json:"password"`
}

func Join(c *gin.Context) {
	ctx := c.Request.Context()
	uriArg := uuidArg{}
	arg := joinArg{}

	// bind uri
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}

	if err := srv.Join(c, uriArg.UUID, arg.Password); err != nil {
		log.For(ctx).Error("join group fail", zap.Error(err), zap.String("uuid", uriArg.UUID))
		return
	}

	c.JSON(http.StatusOK, nil)
}
