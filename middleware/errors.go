package middleware

import (
	"fmt"

	"golang.org/x/text/language"

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
	case "gte":
		return fmt.Sprintf(tip.ValidateGteTip.String(), word, e.Param)
	case "gtfield":
		return fmt.Sprintf(tip.ValidateMinTip.String(), word, e.Param)
	case "oneof":
		return fmt.Sprintf(tip.OneOfTip.String(), word, e.Param)
	case "requiredwhenfield":
		return fmt.Sprintf(tip.RequiredWhenFieldNotEmptyTip.String(), e.Param, word)

	// excludes
	case "excludesrune":
		fallthrough
	case "excludesall":
		fallthrough
	case "excludes":
		return fmt.Sprintf(tip.ExcludeTip.String(), word, e.Param)
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
						writeResponse(c, e, nil)
					}
				case gin.ErrorTypeBind:
					errs := e.Err.(validator.ValidationErrors)
					list := make(map[string]string)
					for _, err := range errs {
						// important, should lower first character
						field := utils.Lower1stCharacter(err.Field)
						list[field] = validationErrorToText(err)
					}
					_ = e.SetMeta(kerror.ErrArgValidateFail)

					writeResponse(c, e, list)
				}

			}

			// If there was no public or bind error, display default 500 message
			if !c.Writer.Written() {
				e := &gin.Error{}
				e = e.SetMeta(kerror.ErrInternalServerErrorGeneral)
				writeResponse(c, e, nil)
			}

		}
	}
}

func writeResponse(c *gin.Context, err *gin.Error, extra interface{}) {
	res, ok := err.Meta.(kerror.ErrResponse)
	if !ok {
		res = kerror.ErrResponse{
			HttpStatus: c.Writer.Status(),
			Code:       c.Writer.Status() * 100, // unknown code,eg: 500 --> 50000
			Tip: tip.Tip{
				language.English.String(): err.Error(),
			},
		}
	}

	// write extra only when the extra of res is nil
	if res.Extra == nil {
		res.Extra = extra
	}

	c.JSON(res.HttpStatus, gin.H{"error": res})
}
