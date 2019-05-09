package main

import (
	"net/http"

	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/si9ma/KillOJ-backend/middleware"

	"github.com/si9ma/KillOJ-backend/api"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	// config gin
	r := gin.Default()

	// tracing
	r.Use(ginhttp.Middleware(gbl.Tracer))

	// error handle
	r.Use(middleware.Errors())

	api.SetupAuthRouter(r) // auth

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}
