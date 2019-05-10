package api

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/gorilla/sessions"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/si9ma/KillOJ-backend/config"
	"github.com/si9ma/KillOJ-backend/gbl"
	"github.com/si9ma/KillOJ-backend/kerror"
	"github.com/si9ma/KillOJ-common/mysql"
	"github.com/si9ma/KillOJ-common/utils"
	otgorm "github.com/smacker/opentracing-gorm"
	"si9ma.com/demo/generated/model"

	"github.com/si9ma/KillOJ-common/log"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/si9ma/KillOJ-common/constants"
)

var supportProvider = []string{"github"}

type auth struct {
	key    string
	secret string
}

func Setup3rdAuth(r *gin.Engine, cfg config.AppConfig) {
	// use goauth,
	// repo : https://github.com/markbates/goth
	useGoAuth(r, cfg)
}

func useGoAuth(r *gin.Engine, cfg config.AppConfig) {
	// set up session
	key := os.Getenv(constants.EnvSessionSecret)
	if key == "" {
		log.Bg().Fatal("Please define environment", zap.String("env", constants.EnvSessionSecret))
	}

	maxAge := 24 * 3600 // 1 days
	isProd := false     // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd
	gothic.Store = store

	goth.UseProviders(
		github.New(os.Getenv(constants.EnvGithubAuthKey), os.Getenv(constants.EnvGithubAuthSecret), utils.GetUrlRoot(cfg.Port)+"/auth/github/callback"), // github
	)
	r.GET("/auth/:provider/callback", jwtMiddleware.LoginHandler) // integration 3rd auth to jwt
	r.GET("/auth/:provider", func(c *gin.Context) {
		ctx := c.Request.Context()
		provider := c.Param("provider")

		if !utils.ContainsString(supportProvider, provider) {
			log.For(ctx).Error("provider is not supported", zap.String("provider", provider))

			_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
				SetMeta(kerror.ErrNotSupportProvider.WithArgs(provider))
			return
		}

		// Compatible with goauth
		ctxWithProvider := context.WithValue(c.Request.Context(), "provider", provider)

		// try to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request.WithContext(ctxWithProvider)); err == nil {
			log.For(ctx).Info("provider already auth", zap.String("provider", provider), zap.Error(err),
				zap.String("userId", gothUser.Name))

			// redirect to callback
			redirectTo := fmt.Sprintf("/auth/%s/callback", provider)
			c.Redirect(http.StatusPermanentRedirect, redirectTo) // integration 3rd auth to jwt
			return
		} else {
			log.For(ctx).Info("auth provider", zap.String("provider", provider))
			gothic.BeginAuthHandler(c.Writer, c.Request.WithContext(ctxWithProvider)) // auth github
			return
		}
	})
}

// third party auth
func thirdAuthenticate(c *gin.Context) (interface{}, error) {
	ctx := c.Request.Context()
	provider := c.Param("provider")
	db := otgorm.SetSpanToGorm(ctx, gbl.DB)

	if !utils.ContainsString(supportProvider, provider) {
		log.For(ctx).Error("provider is not supported", zap.String("provider", provider))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotSupportProvider.WithArgs(provider))
		return "", jwt.ErrFailedAuthentication
	}

	// Compatible with goauth
	ctxWithProvider := context.WithValue(c.Request.Context(), "provider", provider)

	u, err := gothic.CompleteUserAuth(c.Writer, c.Request.WithContext(ctxWithProvider))
	if err != nil {
		log.For(ctx).Error("auth user fail", zap.String("provider", provider), zap.Error(err))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.Err3rdAuthFail.WithArgs(provider))
		return "", jwt.ErrFailedAuthentication
	}

	// github
	user := model.User{}
	if provider == "github" {
		err := db.Where("github_user_id = ?", u.UserID).First(&user).Error
		if hasErr, isNotFound := mysql.ApplyDBError(c, err); isNotFound {
			log.For(ctx).Error("user not signup", zap.String("provider", provider))

			resp := authUserInfo{
				Provider: u.Provider,
				Name:     u.Name,
				UserID:   u.UserID,
			}
			_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
				SetMeta(kerror.ErrNoSignUp.With(resp))
			return "", jwt.ErrFailedAuthentication
		} else if hasErr {
			log.For(ctx).Error("query user by github_user_id fail", zap.String("github_user_id", u.UserID))
			return "", jwt.ErrFailedAuthentication
		}
		return user, nil
	}

	return "", jwt.ErrFailedAuthentication
}

type authUserInfo struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	UserID   string `json:"userID"`
}
