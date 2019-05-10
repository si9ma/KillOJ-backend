package api

import (
	"net/http"
	"testing"

	"github.com/si9ma/KillOJ-backend/config"

	"github.com/gin-gonic/gin"
)

func TestAPI(t *testing.T) {
	r, err := setupRouter()
	if err != nil {
		t.Fatal("setup router fail", err)
	}
	cfg := config.AppConfig{
		Port: "8889",
	}

	SetupAuth(r)
	Setup3rdAuth(r, cfg)

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("/signup", Signup)
	if err := r.Run(":8889"); err != nil {
		t.Fatal(err)
	}
}
