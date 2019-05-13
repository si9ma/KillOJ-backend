package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/srv"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/gin-gonic/gin"
)

func SetupTag(r *gin.Engine) {
	// everyone can access
	r.GET("/tags", GetAllTags)
}

func GetAllTags(c *gin.Context) {
	var err error
	ctx := c.Request.Context()

	tags, err := srv.GetAllTags(c)
	if err != nil {
		log.For(ctx).Error("get tags fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, tags)
}
