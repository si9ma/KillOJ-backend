package main

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/middleware"

	"github.com/si9ma/KillOJ-backend/api"

	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	// config gin
	r := gin.Default()

	// error handle
	r.Use(middleware.Errors())

	// tracing
	r.Use(ginhttp.Middleware(gbl.Tracer))

	api.SetupAuthRouter(r) // auth

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}
