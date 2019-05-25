package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/srv"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/gin-gonic/gin"
)

func SetupTemplate(r *gin.Engine) {
	// everyone can access
	r.GET("/templates", GetAllTemplates)
}

func GetAllTemplates(c *gin.Context) {
	var err error
	ctx := c.Request.Context()

	templates, err := srv.GetAllTemplates(c)
	if err != nil {
		log.For(ctx).Error("get templates fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, templates)
}
