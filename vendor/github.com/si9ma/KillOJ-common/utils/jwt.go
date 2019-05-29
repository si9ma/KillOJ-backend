package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-common/constants"
	"github.com/si9ma/KillOJ-common/model"
)

func SafeGetUserFromJWT(c *gin.Context) (model.User, bool) {
	// get user from jwt
	u, _ := c.Get(constants.JwtIdentityKey)
	user, ok := u.(model.User)
	return user, ok
}
