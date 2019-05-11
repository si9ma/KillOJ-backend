package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/kerror"
	"github.com/si9ma/KillOJ-common/log"
	"go.uber.org/zap"
)

// this middleware should use with gin-jwt
func AuthorizateFunc(handle gin.HandlerFunc, roles ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		user := auth.GetUserFromJWT(c)

		match := false
		for _, role := range roles {
			if role == auth.Role(user.Role) {
				match = true
				break
			}
		}

		// not any role match
		if !match {
			log.For(ctx).Error("access is forbidden", zap.Int("userId", user.ID),
				zap.Int("role", user.Role), zap.Any("need_roles", roles))

			_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
				SetMeta(kerror.ErrForbiddenGeneral)
			return
		}

		handle(c)
	}
}
