package api

import (
	"net/http"
	"strconv"

	"github.com/si9ma/KillOJ-backend/dao"

	"github.com/si9ma/KillOJ-backend/middleware"

	"github.com/si9ma/KillOJ-backend/auth"

	"github.com/jinzhu/gorm"

	"gopkg.in/go-playground/validator.v8"

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
	auth.AuthGroup.POST(ProfilePath, UserInfoEdit)
	auth.AuthGroup.GET(ProfilePath, GetUserInfo)
	auth.AuthGroup.GET("/user/:id", GetOtherUserInfo)
	auth.AuthGroup.PUT("/admin/maintainers/:id",
		middleware.AuthorizateFunc(UpdateMaintainer, auth.Administrator))
	auth.AuthGroup.GET("/admin/maintainers",
		middleware.AuthorizateFunc(GetAllMainters, auth.Administrator))
}

func extractUser(c *gin.Context) (*model.User, bool) {
	ctx := c.Request.Context()
	user := model.User{}

	// bind
	if err := c.ShouldBind(&user); err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			_ = c.Error(err).SetType(gin.ErrorTypeBind)
		}

		_ = c.Error(err).SetType(gin.ErrorTypePublic)
		c.Status(http.StatusBadRequest)
		return nil, false
	}

	// validate argument password,
	// password should't empty when sign up
	if c.Request.RequestURI == SignUpPath {
		if user.Passwd == "" {
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
		if hasErr, _ := mysql.ApplyDBError(c, err); hasErr {
			log.For(ctx).Error("query newUser fail", zap.Error(err), zap.Int("userId", userID))
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
	if newUser.Passwd != "" {
		if newUser.EncryptedPasswd, err = passlib.Hash(newUser.Passwd); err != nil {
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
		if err := db.Model(&oldUser).Updates(newUser).Error; err != nil {
			log.For(ctx).Error("update newUser fail", zap.Error(err))
			_ = c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		log.For(ctx).Info("update newUser success", zap.Int("userId", newUser.ID))
	// sign up newUser
	case SignUpPath:
		newUser.Role = int(auth.Normal) // default user role is normal
		// save
		if err := db.Create(&newUser).Error; err != nil {
			log.For(ctx).Error("create newUser fail", zap.Error(err))
			_ = c.Error(err).SetType(gin.ErrorTypePublic)
			c.Status(http.StatusInternalServerError)
			return
		}
		log.For(ctx).Info("create newUser success", zap.Int("userId", newUser.ID))
	}

	// newUser password and github_user_id should not return,
	// clear newUser password and github_user_id in newUser struct,
	newUser.Passwd = ""
	newUser.GithubUserID = ""
	c.JSON(http.StatusOK, newUser)
}

func GetUserInfo(c *gin.Context) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	userID := auth.GetUserFromJWT(c).ID
	user := model.User{}

	err := db.First(&user, userID).Error
	if hasErr, _ := mysql.ApplyDBError(c, err); hasErr {
		log.For(ctx).Error("get user info fail", zap.Error(err), zap.Int("userId", userID))
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetOtherUserInfo(c *gin.Context) {
	ctx := c.Request.Context()

	// parse id
	userIdStr := c.Param("id")
	userID, err := strconv.Atoi(userIdStr)
	if err != nil {
		log.For(ctx).Error("parse userId fail", zap.Error(err), zap.String("userIdStr", userIdStr))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFoundGeneral)
		return
	}

	user, ok := dao.IsUserExist(c, userID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, user)
}

type updateMaintainerArg struct {
	Role int `json:"role" binding:"exists"`
}

func UpdateMaintainer(c *gin.Context) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	arg := updateMaintainerArg{}

	// parse id
	userIdStr := c.Param("id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		log.For(ctx).Error("parse userId fail", zap.Error(err), zap.String("userIdStr", userIdStr))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFoundGeneral)
		return
	}

	// bind
	if err := c.ShouldBind(&arg); err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			_ = c.Error(err).SetType(gin.ErrorTypeBind)
		}

		_ = c.Error(err).SetType(gin.ErrorTypePublic)
		c.Status(http.StatusBadRequest)
		return
	}

	// arguments check
	if auth.Role(arg.Role) != auth.Normal && auth.Role(arg.Role) != auth.Maintainer {
		log.For(ctx).Error("only allow update maintainer role between normal and maintainer",
			zap.Any("arg", arg.Role))

		fields := map[string]interface{}{
			"role": "1 or 2",
		}
		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrArgValidateFail.With(fields))
		return
	}

	// can't change your self type,
	// because you are administrator
	if userId == auth.GetUserFromJWT(c).ID {
		log.For(ctx).Error("you can't change your self type", zap.Int("userId", userId))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrShouldNotUpdateSelf.WithArgs("type"))
		return
	}

	// if user not exist, return
	if _, ok := dao.IsUserExist(c, userId); !ok {
		return
	}

	user := model.User{ID: userId}
	err = db.Model(&user).Update("role", arg.Role).Error
	if hasErr, isNotFound := mysql.ApplyDBError(c, err); isNotFound {
		log.For(ctx).Error("user not exist", zap.Error(err), zap.Int("userId", userId))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFound.WithArgs(userId))
		return
	} else if hasErr {
		log.For(ctx).Error("update user type fail", zap.Error(err), zap.Int("userid", userId))
		return
	}
}

func GetAllMainters(c *gin.Context) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	var users []model.User
	err := db.Where("role = ?", auth.Maintainer).Find(&users).Error
	if hasErr, _ := mysql.ApplyDBError(c, err); hasErr {
		log.For(ctx).Error("update user type fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, users)
}
