package wrap

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-common/log"
	"gopkg.in/go-playground/validator.v8"
)

func ShouldBind(c *gin.Context, obj interface{}, BindUri bool) (ok bool) {
	var err error

	ctx := c.Request.Context()

	// bind
	if BindUri {
		err = c.ShouldBindUri(obj)
	} else {
		err = c.ShouldBind(obj)
	}
	if err != nil {
		log.For(ctx).Error("bind arguments fail", zap.Error(err),
			zap.Any("obj", obj))

		if _, ok := err.(validator.ValidationErrors); ok {
			_ = c.Error(err).SetType(gin.ErrorTypeBind)
		}

		_ = c.Error(err).SetType(gin.ErrorTypePublic)
		c.Status(http.StatusBadRequest)
		return false
	}

	return true
}
