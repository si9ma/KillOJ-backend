package api

import (
	"testing"

	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
	"github.com/si9ma/KillOJ-backend/gbl"
	"github.com/si9ma/KillOJ-common/tracing"

	"github.com/si9ma/KillOJ-backend/middleware"

	"github.com/gin-gonic/gin"
)

func TestSignup(t *testing.T) {
	r := gin.Default()

	// init tracer
	gbl.Tracer, gbl.TracerCloser = tracing.NewTracer("backend")
	opentracing.SetGlobalTracer(gbl.Tracer)

	r.Use(ginhttp.Middleware(gbl.Tracer))
	r.Use(middleware.Errors())

	r.POST("/signup", Signup)
	if err := r.Run(":8889"); err != nil {
		t.Fatal(err)
	}
}
