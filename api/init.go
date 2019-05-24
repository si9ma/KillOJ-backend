package api

import "github.com/gin-gonic/gin"

func Setup(r *gin.Engine) {
	SetupCatalog(r) // catalog
	SetupUser(r)    // user
	SetupGroup(r)   // group
	SetupContest(r) // contest
	SetupProblem(r) // problem
	SetupTag(r)     // tag
	SetupTheme(r)   // theme
}
