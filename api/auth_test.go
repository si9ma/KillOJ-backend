package api

import (
	"testing"

	"github.com/si9ma/KillOJ-backend/middleware"

	"github.com/gin-gonic/gin"
)

func TestSignup(t *testing.T) {
	r := gin.Default()

	r.Use(middleware.Errors())

	r.POST("/signup", Signup)
	if err := r.Run(":8889"); err != nil {
		t.Fatal(err)
	}
}
