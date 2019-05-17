package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/markbates/goth/gothic"
	"github.com/si9ma/KillOJ-common/mysql"

	"github.com/si9ma/KillOJ-backend/kerror"

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
	UserName string `json:"username" binding:"required"` // user name or email
	Password string `json:"password" binding:"required"`
}

const (
	NoUseGinJwtError = "NoUseGinJwtError"
)

var (
	AuthGroup     *gin.RouterGroup      // auth group
	jwtMiddleware *jwt.GinJWTMiddleware // jwt middleware
)

func SetupAuth(r *gin.Engine) {
	var err error
	jwtSecret := os.Getenv(constants.EnvJWTSecret)
	if jwtSecret == "" {
		log.Bg().Fatal("Please Define environment", zap.String("env", constants.EnvJWTSecret))
		return
	}

	// the jwt middleware
	jwtMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       constants.ProjectName,
		Key:         []byte(jwtSecret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour * 24 * 7, // 7 day
		IdentityKey: constants.JwtIdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(model.User); ok {
				return jwt.MapClaims{
					constants.JwtIdentityKey: v.ID,
					"role":                   v.Role,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)

			// todo There may be a bug here
			userId := claims[constants.JwtIdentityKey].(float64)
			role := claims["role"].(float64)
			return model.User{
				ID:   int(userId),
				Role: int(role),
			}
		},
		Authenticator: authenticate,
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			ctx := c.Request.Context()

			// clear goauth session
			if err := gothic.Logout(c.Writer, c.Request); err != nil {
				log.Bg().Error("goauth logout fail", zap.Error(err))
			}

			if val, ok := c.Get(NoUseGinJwtError); ok {
				if is, ok := val.(bool); ok && is {
					// use custom error handler
					log.For(ctx).Info("don't use gin-jwt error")
					return
				}
			}

			// use gin-jwt error
			c.JSON(code, gin.H{
				"error": map[string]interface{}{
					"code":    kerror.ErrUnauthorizedGeneral.Code,
					"message": message,
				},
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

	r.POST("/login", jwtMiddleware.LoginHandler)

	// auth group
	AuthGroup = r.Group("")
	AuthGroup.Use(jwtMiddleware.MiddlewareFunc())

	AuthGroup.GET("/logout", func(c *gin.Context) {
		if err := gothic.Logout(c.Writer, c.Request); err != nil {
			log.Bg().Error("goauth logout fail", zap.Error(err))
		}
		c.JSON(http.StatusOK, gin.H{
			"result": "success",
		})
	})

	// Refresh time can be longer than token timeout
	AuthGroup.GET("/auth/refresh_token", jwtMiddleware.RefreshHandler)
}

func authenticate(c *gin.Context) (interface{}, error) {
	c.Set(NoUseGinJwtError, true) // use custom error handle

	if c.Request.RequestURI == "/login" {
		// password authenticate
		return passwdAuthenticate(c)
	} else {
		return thirdAuthenticate(c)
	}
}

func passwdAuthenticate(c *gin.Context) (interface{}, error) {
	var (
		loginVals login
		err       error
	)

	parrentCtx := c.Request.Context()
	span, ctx := opentracing.StartSpanFromContext(parrentCtx, "authenticate")
	defer span.Finish()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// bind
	if !wrap.ShouldBind(c, &loginVals, false) {
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
		// username is user name
		err = db.Where("name = ?", loginVals.UserName).First(&user).Error
	}
	if res := mysql.ErrorHandleAndLog(c, err, false,
		"get user by username(email/name)", loginVals.UserName); res == mysql.NotFound {
		log.For(ctx).Error("user not exist", zap.String("username", loginVals.UserName))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrUserNotExist.WithArgs(loginVals.UserName))

		return "", jwt.ErrFailedAuthentication
	} else if res != mysql.Success {
		return "", jwt.ErrFailedAuthentication
	}

	//  verify password
	if newVal, err := passlib.Verify(password, user.EncryptedPasswd); err != nil {
		log.For(ctx).Error("verify password fail", zap.String("username", loginVals.UserName))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrPasswordWrong)

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

				_ = c.Error(err).SetType(gin.ErrorTypePrivate)

				return "", jwt.ErrFailedAuthentication
			}
			log.For(ctx).Info("renew password success", zap.Error(err),
				zap.Int("id", user.ID))
		}
	}

	log.For(ctx).Info("authenticate user success", zap.String("username", loginVals.UserName))
	return user, nil
}

func GetUserFromJWT(c *gin.Context) model.User {
	// get user from jwt
	u, _ := c.Get(constants.JwtIdentityKey)
	return u.(model.User)
}
