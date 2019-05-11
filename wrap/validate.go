package wrap

import (
	"net/http"
	"strconv"

	"github.com/si9ma/KillOJ-backend/kerror"

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

func ExtractIntFromParam(c *gin.Context, key string) (int, bool) {
	ctx := c.Request.Context()

	// parse id
	valStr := c.Param(key)
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.For(ctx).Error("parse params fail", zap.Error(err), zap.String("key", key))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFoundGeneral)
		return val, false
	}

	return val, true
}
