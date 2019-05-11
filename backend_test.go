package main

import (
	"net/http"
	"testing"

	"github.com/si9ma/KillOJ-backend/api"

	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
	"github.com/si9ma/KillOJ-backend/gbl"
	"github.com/si9ma/KillOJ-backend/middleware"
	"github.com/si9ma/KillOJ-common/mysql"
	"github.com/si9ma/KillOJ-common/tracing"

	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/config"
)

func TestBackEnd(t *testing.T) {
	r, err := setupTestRouter()
	if err != nil {
		t.Fatal("setup router fail", err)
	}
	cfg := config.AppConfig{
		Port: "8889",
	}

	auth.SetupAuth(r)
	auth.Setup3rdAuth(r, cfg)
	api.Setup(r)

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	if err := r.Run(":8889"); err != nil {
		t.Fatal(err)
	}
}

func setupTestRouter() (r *gin.Engine, err error) {
	r = gin.Default()

	// init tracer
	gbl.Tracer, gbl.TracerCloser = tracing.NewTracer("backend")
	opentracing.SetGlobalTracer(gbl.Tracer)

	// init db
	if gbl.DB, err = mysql.GetTestDB(); err != nil {
		return nil, err
	}

	r.Use(ginhttp.Middleware(gbl.Tracer))
	r.Use(middleware.Errors())

	return r, nil
}
