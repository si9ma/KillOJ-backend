package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-common/model"
)

// login and auth

func SetupAuthRouter(r *gin.Engine) {
	r.POST("")
}

func Signup(c *gin.Context) {
	user := model.User{}
	if err := c.ShouldBind(&user); err != nil {
		_ = c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	c.JSON(http.StatusOK, user)
}
