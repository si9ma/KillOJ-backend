package api

import (
	"net/http"

	"github.com/go-redis/redis"
	"github.com/si9ma/KillOJ-backend/kerror"

	"github.com/si9ma/KillOJ-backend/data"

	"github.com/si9ma/KillOJ-backend/srv"

	"github.com/si9ma/KillOJ-backend/auth"
	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/gin-gonic/gin"

	"github.com/si9ma/KillOJ-common/model"
)

func SetupGroup(r *gin.Engine) {
	// need auth
	auth.AuthGroup.GET("/groups", GetAllGroups)
	auth.AuthGroup.GET("/groups/group/:id", GetGroup)
	auth.AuthGroup.POST("/groups", AddGroup)
	auth.AuthGroup.PUT("/groups/group/:id", UpdateGroup)
	auth.AuthGroup.POST("/groups/group/:id/invite", Invite2Group)
	auth.AuthGroup.GET("/groups/group/:id/invite", GetGroupInviteInfo)
	auth.AuthGroup.GET("/groups/join/:uuid", JoinGroupQuery)
	auth.AuthGroup.POST("/groups/join/:uuid", JoinGroup)
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

func GetGroupInviteInfo(c *gin.Context) {
	ctx := c.Request.Context()
	uriArg := QueryArg{}

	// bind uri
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	inviteData, err := srv.GetGroupInviteInfo(c, uriArg.ID)
	switch err {
	case redis.Nil: // not found
		log.For(ctx).Error("group invite not exist", zap.Int("groupID", uriArg.ID))
		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFoundOrOutOfDate)
		return
	case nil: // success
		break // continue
	default: // system error
		return
	}

	c.JSON(http.StatusOK, inviteData)
}

func Invite2Group(c *gin.Context) {
	ctx := c.Request.Context()
	inviteData := data.GroupInviteData{}
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
	if err := srv.Invite2Group(c, &inviteData); err != nil {
		log.For(ctx).Error("group invite fail", zap.Error(err), zap.Int("groupId", uriArg.ID))
		return
	}

	// clear password before return
	inviteData.Password = ""
	c.JSON(http.StatusOK, inviteData)
}

func JoinGroupQuery(c *gin.Context) {
	ctx := c.Request.Context()
	uriArg := uuidArg{}

	// bind uri
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	resp, err := srv.JoinGroupQuery(c, uriArg.UUID)
	if err != nil {
		log.For(ctx).Error("query before join fail", zap.Error(err), zap.String("uuid", uriArg.UUID))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func JoinGroup(c *gin.Context) {
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

	if err := srv.JoinGroup(c, uriArg.UUID, arg.Password); err != nil {
		log.For(ctx).Error("join group fail", zap.Error(err), zap.String("uuid", uriArg.UUID))
		return
	}

	c.JSON(http.StatusOK, nil)
}
