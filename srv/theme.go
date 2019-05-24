package srv

import (
	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/gbl"
	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/model"
	"github.com/si9ma/KillOJ-common/mysql"
	otgrom "github.com/smacker/opentracing-gorm"
	"go.uber.org/zap"
)

func SaveTheme(c *gin.Context, theme *model.Theme) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	myID := auth.GetUserFromJWT(c).ID
	theme.UserID = myID
	oldTheme := model.Theme{}

	err := db.Where("user_id = ?", myID).First(&oldTheme).Error
	r := mysql.ErrorHandleAndLog(c, err, false,
		"check if theme exist", myID)
	switch r {
	case mysql.Success:
		theme.ID = oldTheme.ID // if exist, update it
	case mysql.DB_ERROR:
		return err
	}

	err = db.Save(theme).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"save theme", myID) != mysql.Success {
		return err
	}

	log.For(ctx).Info("save theme successful",
		zap.Int("userID", myID))
	return nil
}
