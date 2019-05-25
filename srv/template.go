package srv

import (
	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/mysql"

	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/gin-gonic/gin"

	"github.com/si9ma/KillOJ-common/model"
	otgrom "github.com/smacker/opentracing-gorm"
)

func GetAllTemplates(c *gin.Context) ([]model.Template, error) {
	var err error
	var templates []model.Template

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// error handle
	err = db.Find(&templates).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get templates", nil) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get templates")
	return templates, nil
}