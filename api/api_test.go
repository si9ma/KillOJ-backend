package api

import "testing"

func TestAPI(t *testing.T) {
	r, err := setupRouter()
	if err != nil {
		t.Fatal("setup router fail", err)
	}

	SetupAuth(r)

	r.POST("/signup", Signup)
	if err := r.Run(":8889"); err != nil {
		t.Fatal(err)
	}
}
