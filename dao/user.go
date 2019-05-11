package dao

import (
	"github.com/si9ma/KillOJ-backend/gbl"
	"github.com/si9ma/KillOJ-common/model"

	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-backend/kerror"
	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/mysql"
	otgrom "github.com/smacker/opentracing-gorm"
	"go.uber.org/zap"
)

func IsUserExist(c *gin.Context, userID int) (model.User, bool) {
	user := model.User{}
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	err := db.First(&user, userID).Error
	if hasErr, isNotExist := mysql.ApplyDBError(c, err); isNotExist {
		log.For(ctx).Error("user not exist", zap.Error(err), zap.Int("userId", userID))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFound.WithArgs(userID))
		return user, false
	} else if hasErr {
		log.For(ctx).Error("get user info fail", zap.Error(err), zap.Int("userId", userID))
		return user, false
	}

	return user, true
}
