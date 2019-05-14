package wrap

import (
	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-backend/kerror"
)

func SetInternalServerError(c *gin.Context, err error) {
	if err != nil {
		_ = c.Error(err).SetType(gin.ErrorTypePublic).SetMeta(kerror.ErrInternalServerErrorGeneral)
	}
}

func DiscardGinError(c *gin.Context) {
	c.Errors = []*gin.Error{}
}
