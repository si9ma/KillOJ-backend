package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/si9ma/KillOJ-backend/middleware"

	"github.com/si9ma/KillOJ-backend/auth"

	"github.com/jinzhu/gorm"

	"github.com/si9ma/KillOJ-common/utils"

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
const (
	SignUpPath  = "/signup"
	ProfilePath = "/profile"
)

func SetupUser(r *gin.Engine) {
	r.POST(SignUpPath, UserInfoEdit)

	// should auth
	auth.AuthGroup.PUT(ProfilePath, UserInfoEdit)
	auth.AuthGroup.GET(ProfilePath, GetUserInfo)
	auth.AuthGroup.GET("/user/:id", GetOtherUserInfo)
	auth.AuthGroup.PUT("/admin/maintainers/:id",
		middleware.AuthorizateFunc(UpdateMaintainer, model.Administrator))
	auth.AuthGroup.GET("/admin/maintainers",
		middleware.AuthorizateFunc(GetAllMaintainers, model.Administrator))
}

func extractUser(c *gin.Context) (*model.User, bool) {
	ctx := c.Request.Context()
	user := model.User{}

	if !wrap.ShouldBind(c, &user, false) {
		return nil, false
	}

	// validate argument password,
	// password should't empty when sign up
	if c.Request.RequestURI == SignUpPath {
		if user.Password == "" {
			log.For(ctx).Error("password shouldn't empty when signup")

			// set error
			fields := map[string]string{
				"password": tip.MustNotEmptyTip.String(),
			}
			_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).SetMeta(kerror.ErrArgValidateFail.With(fields))
			return nil, false
		}
	}

	// validate argument organization
	if user.NoInOrganization != "" && user.Organization == "" {
		log.For(ctx).Error("organization should not nil when no_in_organization is not nil",
			zap.String("NoInOrganization", user.NoInOrganization))

		// set error
		fields := map[string]string{
			"no_in_organization": tip.OrgShouldExistWhenNoExistTip.String(),
		}
		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).SetMeta(kerror.ErrArgValidateFail.With(fields))
		return nil, false
	}

	// validate argument github
	if !utils.BothZeroOrNot(user.GithubName, user.GithubUserID) {
		log.For(ctx).Error("github_name and github_user_id should both exist or both not exist",
			zap.String("github_name", user.GithubName), zap.String("github_user_id", user.GithubUserID))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrShouldBothExistOrNot.WithArgs("github_name", "github_user_id"))
		return nil, false
	}

	return &user, true
}

func UserInfoEdit(c *gin.Context) {
	var err error
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	newUser, ok := extractUser(c)
	if !ok {
		return
	}

	oldUser := model.User{}
	// when update user info
	if c.Request.RequestURI == ProfilePath {
		userID := auth.GetUserFromJWT(c).ID // get ID from jwt
		err := db.First(&oldUser, userID).Error
		if mysql.ErrorHandleAndLog(c, err, true,
			"query user", userID) != mysql.Success {
			return
		}
	}

	// unique check
	// new value -- old value
	uniqueCheckList := []map[string]mysql.ValuePair{
		{
			"organization": mysql.ValuePair{
				NewVal: newUser.Organization,
				OldVal: oldUser.Organization,
			},
			"no_in_organization": mysql.ValuePair{
				NewVal: newUser.NoInOrganization,
				OldVal: oldUser.NoInOrganization,
			},
		},
		{
			"email": mysql.ValuePair{
				NewVal: newUser.Email,
				OldVal: oldUser.Email,
			},
		},
		{
			"nick_name": mysql.ValuePair{
				NewVal: newUser.NickName,
				OldVal: oldUser.NickName,
			},
		},
		{
			"github_user_id": mysql.ValuePair{
				NewVal: newUser.GithubUserID,
				OldVal: oldUser.GithubUserID,
			},
		},
		{
			"github_name": mysql.ValuePair{
				NewVal: newUser.GithubName,
				OldVal: oldUser.GithubName,
			},
		},
	}

	// check
	for _, checkMap := range uniqueCheckList {
		if !mysql.ShouldUnique(c, ctx, db, checkMap, func(db *gorm.DB) error {
			return db.First(&model.User{}).Error
		}) {
			return
		}
	}

	// encrypt password,
	// only when password not empty
	if newUser.Password != "" {
		if newUser.EncryptedPasswd, err = passlib.Hash(newUser.Password); err != nil {
			log.For(ctx).Error("encrypt password fail", zap.Error(err))
			_ = c.Error(err).SetType(gin.ErrorTypePrivate)
			c.Status(http.StatusInternalServerError)
			return
		}
		log.For(ctx).Info("encrypt password success")
	}

	switch c.Request.RequestURI {
	// update newUser info
	case ProfilePath:
		// these field shouldn't update
		newUser.ID = oldUser.ID
		newUser.GithubName = oldUser.GithubName
		newUser.GithubUserID = oldUser.GithubUserID

		// update
		err := db.Model(&oldUser).Updates(newUser).Error
		if mysql.ErrorHandleAndLog(c, err, true,
			"update user", newUser.ID) != mysql.Success {
			return
		}
		log.For(ctx).Info("update newUser success", zap.Int("userId", newUser.ID))
	// sign up newUser
	case SignUpPath:
		newUser.Role = int(model.Normal) // default user role is normal
		// save
		if err := db.Create(&newUser).Error; err != nil {
			log.For(ctx).Error("create newUser fail", zap.Error(err))
			_ = c.Error(err).SetType(gin.ErrorTypePublic)
			c.Status(http.StatusInternalServerError)
			return
		}
		log.For(ctx).Info("create newUser success", zap.Int("userId", newUser.ID))
	}

	c.JSON(http.StatusOK, newUser)
}

func GetUserInfo(c *gin.Context) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	userID := auth.GetUserFromJWT(c).ID
	user := model.User{}

	err := db.First(&user, userID).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get user info", userID) != mysql.Success {
		return
	}

	log.For(ctx).Info("get user info success", zap.Int("userId", user.ID))
	c.JSON(http.StatusOK, user)
}

func GetOtherUserInfo(c *gin.Context) {
	ctx := c.Request.Context()
	arg := QueryArg{}
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// bind uri
	if !wrap.ShouldBind(c, &arg, true) {
		return
	}

	user := model.User{}
	err := db.First(&user, arg.ID).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get user info", arg.ID) != mysql.Success {
		return
	}

	log.For(ctx).Info("get user info success", zap.Int("userId", user.ID))
	c.JSON(http.StatusOK, user)
}

type updateMaintainerArg struct {
	Role int `json:"role" binding:"exists,oneof=1 2"`
}

func UpdateMaintainer(c *gin.Context) {
	var err error

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	arg := updateMaintainerArg{}
	uriArg := QueryArg{}

	// bind url argument
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind other argument
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}

	// can't change your self role,
	// because you are administrator
	if uriArg.ID == auth.GetUserFromJWT(c).ID {
		log.For(ctx).Error("you can't change your self role", zap.Int("userId", uriArg.ID))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrShouldNotUpdateSelf.WithArgs("role"))
		return
	}

	// if user not exist, return
	err = db.First(&model.User{}).Error
	if mysql.ErrorHandleAndLog(c, err, true, "get user", uriArg.ID) != mysql.Success {
		return
	}

	user := model.User{ID: uriArg.ID}
	err = db.Model(&user).Update("role", arg.Role).Error
	if res := mysql.ErrorHandleAndLog(c, err, false,
		"update maintainer role", uriArg.ID); res == mysql.NotFound {
		log.For(ctx).Error("user not exist", zap.Error(err), zap.Int("userId", uriArg.ID))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFound.WithArgs(uriArg.ID))
		return
	} else if res != mysql.Success {
		return
	}

	log.For(ctx).Info("update maintainer role success", zap.Int("userId", user.ID))
}

func GetAllMaintainers(c *gin.Context) {
	var err error

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	arg := PageArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}
	offset := (arg.Page - 1) * arg.PageSize

	var users []model.User
	if arg.Order != "" {
		err = db.Where("role = ?", model.Maintainer).Order(arg.Order).Offset(offset).Limit(arg.PageSize).Find(&users).Error
	} else {
		err = db.Where("role = ?", model.Maintainer).Offset(offset).Limit(arg.PageSize).Find(&users).Error
	}
	if mysql.ErrorHandleAndLog(c, err, true, "get maintainers", nil) != mysql.Success {
		return
	}

	log.For(ctx).Info("success get maintainers")
	c.JSON(http.StatusOK, users)
}
