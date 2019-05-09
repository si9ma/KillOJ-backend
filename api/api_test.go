package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPI(t *testing.T) {
	r, err := setupRouter()
	if err != nil {
		t.Fatal("setup router fail", err)
	}

	SetupAuth(r)

	// Ping test
	r.GET("/auth/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("/signup", Signup)
	if err := r.Run(":8889"); err != nil {
		t.Fatal(err)
	}
}
