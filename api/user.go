package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-common/mysql"

	"github.com/si9ma/KillOJ-common/tip"

	"github.com/si9ma/KillOJ-backend/kerror"

	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/si9ma/KillOJ-common/log"
	"go.uber.org/zap"

	otgrom "github.com/smacker/opentracing-gorm"
	"gopkg.in/hlandau/passlib.v1"

	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-common/model"
)

// login and auth

func SetupAuthRouter(r *gin.Engine) {
	r.POST("/signup", Signup)
}

func Signup(c *gin.Context) {
	var err error
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	user := model.User{}

	// bind
	if err := c.ShouldBind(&user); err != nil {
		_ = c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	// validate argument
	if user.NoInOrganization != "" && user.Organization == "" {
		log.For(ctx).Error("organization should not nil when no_in_organization is not nil",
			zap.String("NoInOrganization", user.NoInOrganization))

		// set error
		fields := map[string]string{
			"no_in_organization": tip.OrgShouldExistWhenNoExistTip.String(),
		}
		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).SetMeta(kerror.ErrArgValidateFail.With(fields))
		return
	}

	// Organization and NoInOrganization must unique
	if user.NoInOrganization != "" && user.Organization != "" {
		tmpUser := model.User{}
		err := db.Where("organization = ? AND no_in_organization = ?",
			user.Organization, user.NoInOrganization).First(&tmpUser).Error

		if hasErr, isNotFound := mysql.ApplyDBError(c, err); !hasErr {
			// already exist
			log.For(ctx).Error("NoInOrganization already exist",
				zap.String("NoInOrganization", user.NoInOrganization),
				zap.String("organization", user.Organization))

			_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
				SetMeta(kerror.ErrUserAlreadyExistInOrg.WithArgs(user.NoInOrganization, user.Organization))
			return
		} else if !isNotFound {
			log.For(ctx).Error("query user by organization and no_in_organization fail", zap.Error(err),
				zap.String("organization", user.Organization), zap.String("no_in_organization", user.NoInOrganization))
			return
		}
	}

	// email should unique
	tmpUser := model.User{}
	err = db.Where("email = ?", user.Email).First(&tmpUser).Error
	if hasErr, isNotFound := mysql.ApplyDBError(c, err); !hasErr {
		// already exist
		log.For(ctx).Error("email already exist", zap.String("email", user.Email))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrAlreadyExist.WithArgs(user.Email))
		return
	} else if !isNotFound {
		log.For(ctx).Error("query user by email fail", zap.Error(err),
			zap.String("email", user.Email))
		return
	}

	// nick name should unique
	tmpUser = model.User{}
	err = db.Where("nick_name = ?", user.NickName).First(&tmpUser).Error
	if hasErr, isNotFound := mysql.ApplyDBError(c, err); !hasErr {
		// already exist
		log.For(ctx).Error("nick name already exist", zap.String("nick_name", user.NickName))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrAlreadyExist.WithArgs(user.NickName))
		return
	} else if !isNotFound {
		log.For(ctx).Error("query user by nick name fail", zap.Error(err),
			zap.String("email", user.NickName))
		return
	}

	// encrypt password
	if user.EncryptedPasswd, err = passlib.Hash(user.Passwd); err != nil {
		log.For(ctx).Error("encrypt password fail", zap.Error(err))
		_ = c.Error(err).SetType(gin.ErrorTypePrivate)
		c.Status(http.StatusInternalServerError)
		return
	}
	log.For(ctx).Info("encrypt password success")

	// save
	if err := db.Create(&user).Error; err != nil {
		log.For(ctx).Error("create user fail", zap.Error(err))
		_ = c.Error(err).SetType(gin.ErrorTypePublic)
		c.Status(http.StatusInternalServerError)
		return
	}
	log.For(ctx).Info("create user success", zap.Int("userId", user.ID))

	// user password should not return,
	// clear user password in user struct,
	user.Passwd = ""
	c.JSON(http.StatusOK, user)
}
