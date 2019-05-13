package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/kerror"

	"github.com/si9ma/KillOJ-common/model"

	"github.com/si9ma/KillOJ-backend/srv"

	"github.com/si9ma/KillOJ-backend/auth"
	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/gin-gonic/gin"
)

func setupProblem(r *gin.Engine) {
	// need auth
	auth.AuthGroup.GET("/problems", GetAllProblems)
	auth.AuthGroup.GET("/problems/problem/:id", GetProblem)
	auth.AuthGroup.POST("/problems", AddProblem)
	auth.AuthGroup.PUT("/problems/problem/:id", UpdateProblem)
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

	problems, err := srv.GetAllProblems(c, arg.Page, arg.PageSize, arg.Order)
	if err != nil {
		log.For(ctx).Error("get problems fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, problems)
}

func GetProblem(c *gin.Context) {
	ctx := c.Request.Context()
	arg := QueryArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, true) {
		return
	}

	problem, err := srv.GetProblem(c, arg.ID)
	if err != nil {
		log.For(ctx).Error("get problem info fail", zap.Error(err), zap.Int("problemId", arg.ID))
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

	// when delete tag, must provide id
	for _, tag := range newProblem.Tags {
		if tag.DeleteIt && tag.ID <= 0 {
			log.For(ctx).Error("must provide id when delete tag")
			_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
				SetMeta(kerror.ErrMustProvideWhenAnotherExist.WithArgs("delete_it of tag", "tag id"))
			return
		}
	}

	// when delete sample, must provide id
	for _, sample := range newProblem.ProblemSamples {
		if sample.DeleteIt && sample.ID <= 0 {
			log.For(ctx).Error("must provide id when delete sample")
			_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
				SetMeta(kerror.ErrMustProvideWhenAnotherExist.WithArgs("delete_it of sample", "tag id"))
			return
		}
	}

	// use id in uri path
	newProblem.ID = uriArg.ID
	if err := srv.UpdateProblem(c, &newProblem); err != nil {
		log.For(ctx).Error("update problem fail", zap.Error(err), zap.Int("problemId", newProblem.ID))
		return
	}

	c.JSON(http.StatusOK, newProblem)
}
