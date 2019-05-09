package api

import (
	"testing"
)

func TestSignup(t *testing.T) {
	r, err := setupRouter()
	if err != nil {
		t.Fatal("setup router fail", err)
	}

	r.POST("/signup", Signup)
	if err := r.Run(":8889"); err != nil {
		t.Fatal(err)
	}
}
