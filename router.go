package main

import (
	"net/http"
	"time"

	"github.com/si9ma/KillOJ-backend/validator"

	"github.com/si9ma/KillOJ-backend/api"
	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/config"

	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/gin-contrib/cors"
	"github.com/si9ma/KillOJ-backend/middleware"

	"github.com/gin-gonic/gin"
)

func setupRouter(cfg *config.Config) *gin.Engine {
	// config gin
	r := gin.Default()

	// gin tracing middleware
	r.Use(ginhttp.Middleware(gbl.Tracer))

	// error handle middleware
	r.Use(middleware.Errors())

	// cors
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// set up auth
	auth.SetupAuth(r)
	auth.Setup3rdAuth(r, cfg.App)

	// setup custom validator
	validator.SetupValidator()

	// set up api
	api.Setup(r)

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}
