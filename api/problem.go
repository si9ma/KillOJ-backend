package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-common/model"

	"github.com/si9ma/KillOJ-backend/srv"

	"github.com/si9ma/KillOJ-backend/auth"
	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/gin-gonic/gin"
)

func SetupProblem(r *gin.Engine) {
	// need auth
	auth.AuthGroup.GET("/problems", GetAllProblems)
	auth.AuthGroup.GET("/problems/problem/:id", GetProblem)
	auth.AuthGroup.POST("/problems", AddProblem)
	auth.AuthGroup.PUT("/problems/problem/:id", UpdateProblem)
	auth.AuthGroup.GET("/problems/problem/:id/vote", VoteProblem)
	//auth.AuthProblem.DELETE("/problems/:id", DeleteProblem)
}

func GetAllProblems(c *gin.Context) {
	var err error
	ctx := c.Request.Context()
	arg := PageArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}

	problems, err := srv.GetAllProblems(c, arg.Page, arg.PageSize, arg.Order, arg.Of, arg.ID)
	if err != nil {
		log.For(ctx).Error("get problems fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, problems)
}

func GetProblem(c *gin.Context) {
	ctx := c.Request.Context()
	uriArg := QueryArg{}
	arg := QueryExtraArg{}

	// bind
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind query params
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}

	problem, err := srv.GetProblem(c, uriArg.ID, arg.ForUpdate)
	if err != nil {
		log.For(ctx).Error("get problem info fail", zap.Error(err), zap.Int("problemId", uriArg.ID))
		return
	}

	c.JSON(http.StatusOK, problem)
}

func AddProblem(c *gin.Context) {
	ctx := c.Request.Context()
	newProblem := model.Problem{}
	userId := auth.GetUserFromJWT(c).ID

	// bind
	if !wrap.ShouldBind(c, &newProblem, false) {
		return
	}

	newProblem.OwnerID = userId
	if err := srv.AddProblem(c, &newProblem); err != nil {
		log.For(ctx).Error("add problem fail", zap.Error(err), zap.Int("problemId", newProblem.ID))
		return
	}

	c.JSON(http.StatusOK, newProblem)
}

func UpdateProblem(c *gin.Context) {
	ctx := c.Request.Context()
	newProblem := model.Problem{}
	uriArg := QueryArg{}

	// bind uri params
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind request params
	if !wrap.ShouldBind(c, &newProblem, false) {
		return
	}

	// use id in uri path
	newProblem.ID = uriArg.ID
	if err := srv.UpdateProblem(c, &newProblem); err != nil {
		log.For(ctx).Error("update problem fail", zap.Error(err), zap.Int("problemId", newProblem.ID))
		return
	}

	c.JSON(http.StatusOK, newProblem)
}

func VoteProblem(c *gin.Context) {
	ctx := c.Request.Context()
	vote := model.UserVoteProblem{}
	uriArg := QueryArg{}

	// bind uri params
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind request params
	if !wrap.ShouldBind(c, &vote, false) {
		return
	}

	if err := srv.VoteProblem(c, uriArg.ID, vote.Attitude); err != nil {
		log.For(ctx).Error("user vote problem", zap.Error(err), zap.Int("problemId", uriArg.ID))
		return
	}

	c.JSON(http.StatusOK, nil)
}
