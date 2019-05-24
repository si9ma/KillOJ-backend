package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/wrap"
	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/model"
	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-backend/auth"

	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-backend/srv"
)

func SetupTheme(r *gin.Engine) {
	auth.AuthGroup.POST("/theme", SaveTheme)
}

func SaveTheme(c *gin.Context) {
	ctx := c.Request.Context()
	theme := model.Theme{}

	if !wrap.ShouldBind(c, &theme, false) {
		return
	}

	if err := srv.SaveTheme(c, &theme); err != nil {
		log.For(ctx).Error("save theme fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, theme)
}
