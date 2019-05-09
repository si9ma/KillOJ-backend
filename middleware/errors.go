package middleware

import (
	"fmt"
	"github.com/si9ma/KillOJ-common/tip"
	"net/http"

	"github.com/si9ma/KillOJ-common/utils"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v8"
)

func ValidationErrorToText(e *validator.FieldError) string {
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

		// Only run if there are some errors to handle
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				// Find out what type of error it is
				switch e.Type {
				case gin.ErrorTypePublic:
					// Only output public errors if nothing has been written yet
					if !c.Writer.Written() {
						c.JSON(c.Writer.Status(), gin.H{"error": e.Error()})
					}
				case gin.ErrorTypeBind:
					errs := e.Err.(validator.ValidationErrors)
					list := make(map[string]string)
					for _, err := range errs {
						// important, should lower first character
						field := utils.Lower1stCharacter(err.Field)
						list[field] = ValidationErrorToText(err)
					}

					// Make sure we maintain the preset response status
					status := http.StatusBadRequest
					if c.Writer.Status() != http.StatusOK {
						status = c.Writer.Status()
					}
					c.JSON(status, gin.H{"error": map[]})

				default:
					// Log all other errors
				}

			}
			// If there was no public or bind error, display default 500 message
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "fk"})
			}
		}
	}
}
