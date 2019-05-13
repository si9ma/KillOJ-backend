package api

import "github.com/gin-gonic/gin"

func Setup(r *gin.Engine) {
	SetupCatalog(r) // catalog
	SetupUser(r)    // user
	setupGroup(r)   // group
	setupContest(r) // contest
	setupProblem(r) // contest
}
