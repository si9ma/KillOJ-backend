package api

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
	"github.com/si9ma/KillOJ-backend/gbl"
	"github.com/si9ma/KillOJ-backend/middleware"
	"github.com/si9ma/KillOJ-common/mysql"
	"github.com/si9ma/KillOJ-common/tracing"
)

func setupRouter() (r *gin.Engine, err error) {
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
