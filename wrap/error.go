package wrap

import (
	"github.com/gin-gonic/gin"
)

func SetInternalServerError(c *gin.Context, err error) {
	if err != nil {
		_ = c.Error(err).SetType(gin.ErrorTypePublic)
	}
}
