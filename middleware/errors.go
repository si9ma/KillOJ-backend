package middleware

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/si9ma/KillOJ-backend/kerror"
	"github.com/si9ma/KillOJ-common/tip"

	"github.com/si9ma/KillOJ-common/utils"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v8"
)

func validationErrorToText(e *validator.FieldError) string {
	word := utils.Lower1stCharacter(e.Field)

	switch e.Tag {
	case "required":
		return fmt.Sprintf(tip.ValidateRequireTip.String(), word)
	case "max":
		return fmt.Sprintf(tip.ValidateMaxTip.String(), word, e.Param)
	case "min":
		return fmt.Sprintf(tip.ValidateMinTip.String(), word, e.Param)
	case "email":
		return fmt.Sprintf(tip.ValidateEmailTip.String())
	case "len":
		return fmt.Sprintf(tip.ValidateLenTip.String(), word, e.Param)
	}
	return fmt.Sprintf(tip.ValidateInvalidTip.String(), word)
}

// This method collects all errors
func Errors() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Next()
		ctx := c.Request.Context()

		// Only run if there are some errors to handle
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				// log
				log.For(ctx).Error("request fail", zap.Error(e))

				// Find out what type of error it is
				switch e.Type {
				case gin.ErrorTypePublic:
					// Only output public errors if nothing has been written yet
					if !c.Writer.Written() {
						c.JSON(c.Writer.Status(), gin.H{"error": generateErrResponse(c, e)})
					}
				case gin.ErrorTypeBind:
					errs := e.Err.(validator.ValidationErrors)
					list := make(map[string]string)
					for _, err := range errs {
						// important, should lower first character
						field := utils.Lower1stCharacter(err.Field)
						list[field] = validationErrorToText(err)
					}

					// Make sure we maintain the preset response status
					status := http.StatusBadRequest
					if c.Writer.Status() != http.StatusOK {
						status = c.Writer.Status()
					}

					c.JSON(status, gin.H{
						"error": map[string]interface{}{
							"code":    kerror.ArgValidateFail,
							"message": tip.ArgValidateFailTip.String(),
							"errors":  list,
						},
					})
				}

			}

			// If there was no public or bind error, display default 500 message
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError,
					gin.H{
						"error": map[string]interface{}{
							"code":    kerror.InternalServerErrorGeneral,
							"message": tip.InternalServerErrorTip.String(),
						},
					},
				)
			}

		}
	}
}

func generateErrResponse(c *gin.Context, err *gin.Error) map[string]interface{} {
	res := make(map[string]interface{})
	code, ok := err.Meta.(kerror.ResponseErrno)
	if !ok {
		code = kerror.ResponseErrno(c.Writer.Status() * 100) // unknown code, eg: 401 --> 40100
	}

	res["code"] = code
	res["message"] = err.Error()
	return res
}
