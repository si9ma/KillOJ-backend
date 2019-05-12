package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/kerror"

	"github.com/go-redis/redis"
	"github.com/si9ma/KillOJ-backend/data"

	"github.com/si9ma/KillOJ-backend/srv"

	"github.com/si9ma/KillOJ-backend/auth"
	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/gin-gonic/gin"

	"github.com/si9ma/KillOJ-common/model"
)

func setupContest(r *gin.Engine) {
	// need auth
	auth.AuthGroup.GET("/contests", GetAllContests)
	auth.AuthGroup.GET("/contests/contest/:id", GetContest)
	auth.AuthGroup.POST("/contests", AddContest)
	auth.AuthGroup.PUT("/contests/contest/:id", UpdateContest)
	auth.AuthGroup.POST("/contests/contest/:id/invite", Invite2Contest)
	auth.AuthGroup.GET("/contests/contest/:id/invite", GetContestInviteInfo)
	auth.AuthGroup.GET("/contests/join/:uuid", JoinContestQuery)
	auth.AuthGroup.POST("/contests/join/:uuid", JoinContest)
	//auth.AuthContest.DELETE("/contests/:id", DeleteContest)
}

func GetAllContests(c *gin.Context) {
	var err error
	ctx := c.Request.Context()
	arg := PageArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}

	contests, err := srv.GetAllContests(c, arg.Page, arg.PageSize, arg.Order)
	if err != nil {
		log.For(ctx).Error("get contests fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, contests)
}

func GetContest(c *gin.Context) {
	ctx := c.Request.Context()
	arg := QueryArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, true) {
		return
	}

	contest, err := srv.GetContest(c, arg.ID)
	if err != nil {
		log.For(ctx).Error("get contest info fail", zap.Error(err), zap.Int("contestId", arg.ID))
		return
	}

	c.JSON(http.StatusOK, contest)
}

func AddContest(c *gin.Context) {
	ctx := c.Request.Context()
	newContest := model.Contest{}
	userId := auth.GetUserFromJWT(c).ID

	// bind
	if !wrap.ShouldBind(c, &newContest, false) {
		return
	}

	newContest.OwnerID = userId
	if err := srv.AddContest(c, &newContest); err != nil {
		log.For(ctx).Error("add contest fail", zap.Error(err), zap.Int("contestId", newContest.ID))
		return
	}

	c.JSON(http.StatusOK, newContest)
}

func UpdateContest(c *gin.Context) {
	ctx := c.Request.Context()
	newContest := model.Contest{}
	uriArg := QueryArg{}

	// bind uri params
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind request params
	if !wrap.ShouldBind(c, &newContest, false) {
		return
	}

	// use id in uri path
	newContest.ID = uriArg.ID
	if err := srv.UpdateContest(c, &newContest); err != nil {
		log.For(ctx).Error("update contest fail", zap.Error(err), zap.Int("contestId", newContest.ID))
		return
	}

	c.JSON(http.StatusOK, newContest)
}

//
//func DeleteContest(c *gin.Context) {
//	ctx := c.Request.Context()
//	uriArg := QueryArg{}
//
//	// bind uri params
//	if !wrap.ShouldBind(c, &uriArg, true) {
//		return
//	}
//
//	if err := srv.DeleteContest(c, uriArg.ID); err != nil {
//		log.For(ctx).Error("delete contest fail", zap.Error(err), zap.Int("contestId", uriArg.ID))
//		return
//	}
//
//	c.JSON(http.StatusOK, nil)
//}

func GetContestInviteInfo(c *gin.Context) {
	ctx := c.Request.Context()
	uriArg := QueryArg{}

	// bind uri
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	inviteData, err := srv.GetContestInviteInfo(c, uriArg.ID)
	switch err {
	case redis.Nil: // not found
		log.For(ctx).Error("contest invite not exist", zap.Int("contestID", uriArg.ID))
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

func Invite2Contest(c *gin.Context) {
	ctx := c.Request.Context()
	inviteData := data.ContestInviteData{}
	uriArg := QueryArg{}

	// bind uri
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind
	if !wrap.ShouldBind(c, &inviteData, false) {
		return
	}

	inviteData.ContestID = uriArg.ID
	if err := srv.Invite2Contest(c, &inviteData); err != nil {
		log.For(ctx).Error("contest invite fail", zap.Error(err), zap.Int("contestId", uriArg.ID))
		return
	}

	// clear password before return
	inviteData.Password = ""
	c.JSON(http.StatusOK, inviteData)
}

func JoinContestQuery(c *gin.Context) {
	ctx := c.Request.Context()
	uriArg := uuidArg{}

	// bind uri
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	resp, err := srv.JoinContestQuery(c, uriArg.UUID)
	if err != nil {
		log.For(ctx).Error("query before join fail", zap.Error(err), zap.String("uuid", uriArg.UUID))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func JoinContest(c *gin.Context) {
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

	if err := srv.JoinContest(c, uriArg.UUID, arg.Password); err != nil {
		log.For(ctx).Error("join contest fail", zap.Error(err), zap.String("uuid", uriArg.UUID))
		return
	}

	c.JSON(http.StatusOK, nil)
}
