package api

import "github.com/gin-gonic/gin"

func Setup(r *gin.Engine) {
	setupCatalog(r) // catalog
	SetupUser(r)    // user
}
