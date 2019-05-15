package main

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/validator"

	"github.com/si9ma/KillOJ-backend/api"
	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/config"

	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/si9ma/KillOJ-backend/gbl"

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
