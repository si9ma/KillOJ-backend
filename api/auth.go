package api

import (
	"os"
	"time"

	"gopkg.in/hlandau/passlib.v1"

	"github.com/si9ma/KillOJ-common/utils"

	"github.com/si9ma/KillOJ-common/log"
	"go.uber.org/zap"

	"github.com/opentracing/opentracing-go"

	"github.com/si9ma/KillOJ-backend/gbl"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-common/constants"
	"github.com/si9ma/KillOJ-common/model"
	otgrom "github.com/smacker/opentracing-gorm"
)

type login struct {
	UserName string `json:"username" binding:"required"` // nick name or email
	Password string `json:"password" binding:"required"`
}

var (
	authGroup   *gin.RouterGroup
	identityKey = "id"
)

func SetupAuth(r *gin.Engine) {
	jwtSecret := os.Getenv(constants.EnvJWTSecret)
	if jwtSecret == "" {
		log.Bg().Fatal("Please Define environment", zap.String("env", constants.EnvJWTSecret))
		return
	}

	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       constants.ProjectName,
		Key:         []byte(jwtSecret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour * 24 * 7, // 7 day
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*model.User); ok {
				return jwt.MapClaims{
					identityKey: v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &model.User{
				ID: claims["id"].(int),
			}
		},
		Authenticator: authenticate,
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value.
		// This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	if err != nil {
		log.Bg().Fatal("init jwt fail", zap.Error(err))
		return
	}

	r.POST("/login", authMiddleware.LoginHandler)

	authGroup = r.Group("/auth")

	// Refresh time can be longer than token timeout
	authGroup.GET("/refresh_token", authMiddleware.RefreshHandler)
	authGroup.Use(authMiddleware.MiddlewareFunc())
}

func authenticate(c *gin.Context) (interface{}, error) {
	var (
		loginVals login
		err       error
	)

	parrentCtx := c.Request.Context()
	span, ctx := opentracing.StartSpanFromContext(parrentCtx, "authenticate")
	defer span.Finish()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// bind
	if err := c.ShouldBind(&loginVals); err != nil {
		log.For(ctx).Error("bind login info fail",
			zap.String("username", loginVals.UserName),
			zap.String("password", loginVals.Password))

		// use gin-jwt error
		//_ = c.Error(err).SetType(gin.ErrorTypeBind)
		return "", jwt.ErrMissingLoginValues
	}
	userName := loginVals.UserName
	password := loginVals.Password

	// query db
	user := model.User{}
	if utils.CheckEmail(userName) {
		// username is email
		err = db.Where("email = ?", loginVals.UserName).First(&user).Error
	} else {
		// username is nick name
		err = db.Where("nick_name = ?", loginVals.UserName).First(&user).Error
	}
	if err != nil {
		log.For(ctx).Error("query user fail", zap.String("username", loginVals.UserName))
		return "", jwt.ErrFailedAuthentication
	}

	if newVal, err := passlib.Verify(password, user.EncryptedPasswd); err != nil {
		log.For(ctx).Error("verify password fail", zap.String("username", loginVals.UserName))
		return "", jwt.ErrFailedAuthentication
	} else {
		// The context has decided, as per its policy, that
		// the hash which was used to validate the password
		// should be changed. It has upgraded the hash using
		// the verified password.
		// refer : https://github.com/hlandau/passlib
		if newVal != "" {
			if err := db.Model(&user).Update("passwd", newVal).Error; err != nil {
				log.For(ctx).Error("renew password fail", zap.Error(err),
					zap.Int("id", user.ID))

				// todo There may be a bug here
				//return "",jwt.ErrFailedAuthentication
			}
			log.For(ctx).Info("renew password success", zap.Error(err),
				zap.Int("id", user.ID))
		}
	}

	log.For(ctx).Info("authenticate user success", zap.String("username", loginVals.UserName))
	return user, nil
}
