// error package for killoj
package kerror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/si9ma/KillOJ-common/utils"

	"github.com/si9ma/KillOJ-common/log"
	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/tip"
)

var (
	EmptyError = errors.New("") // error variable with empty message
)

type ErrResponse struct {
	HttpStatus int
	Code       int
	Tip        tip.Tip
	Extra      interface{}
}

var (
	// 400xx : bad request
	ErrBadRequestGeneral           = ErrResponse{http.StatusBadRequest, 40000, tip.BadRequestGeneralTip, nil}
	ErrArgValidateFail             = ErrResponse{http.StatusBadRequest, 40001, tip.ArgValidateFailTip, nil}
	ErrNotExist                    = ErrResponse{http.StatusBadRequest, 40002, tip.NotExistTip, nil}
	ErrAlreadyExist                = ErrResponse{http.StatusBadRequest, 40003, tip.AlreadyExistTip, nil}
	ErrShouldBothExistOrNot        = ErrResponse{http.StatusBadRequest, 40004, tip.ShouldBothExistOrNotTip, nil}
	ErrShouldNotUpdateSelf         = ErrResponse{http.StatusBadRequest, 40005, tip.ShouldNotUpdateSelfTip, nil}
	ErrAlreadyInvite               = ErrResponse{http.StatusBadRequest, 40006, tip.AlreadyInviteTip, nil}
	ErrAlreadyFinished             = ErrResponse{http.StatusBadRequest, 40007, tip.AlreadyFinishTip, nil}
	ErrMustProvideWhenAnotherExist = ErrResponse{http.StatusBadRequest, 40008, tip.MustProvideWhenAnotherExistTip, nil}
	ErrNotComplete                 = ErrResponse{http.StatusBadRequest, 40009, tip.TaskNotCompleteTip, nil}
	ErrAtLeast                     = ErrResponse{http.StatusBadRequest, 40010, tip.AtLeastTip, nil}

	// 401xx:
	ErrUnauthorizedGeneral = ErrResponse{http.StatusUnauthorized, 40100, tip.UnauthorizedGeneralTip, nil}
	ErrUserNotExist        = ErrResponse{http.StatusUnauthorized, 40101, tip.UserNotExistTip, nil}
	ErrPasswordWrong       = ErrResponse{http.StatusUnauthorized, 40102, tip.PasswordWrong, nil}
	Err3rdAuthFail         = ErrResponse{http.StatusUnauthorized, 40103, tip.ThirdAuthFailTip, nil}
	ErrNoSignUp            = ErrResponse{http.StatusUnauthorized, 40104, tip.NoSignupTip, nil}
	ErrNotSupportProvider  = ErrResponse{http.StatusUnauthorized, 40105, tip.NotSupportProviderTip, nil}
	ErrHaveRunningTask     = ErrResponse{http.StatusUnauthorized, 40106, tip.HaveRunningTaskTip, nil}

	// 403xx : forbidden
	ErrForbiddenGeneral = ErrResponse{http.StatusForbidden, 40300, tip.ForbiddenTip, nil}

	// 404xx : not found
	ErrNotFoundGeneral     = ErrResponse{http.StatusForbidden, 40400, tip.NotFoundTip, nil}
	ErrNotFound            = ErrResponse{http.StatusBadRequest, 40401, tip.NotExistTip, nil}
	ErrNotFoundOrOutOfDate = ErrResponse{http.StatusBadRequest, 40401, tip.NotExistOrOutOfDateTip, nil}

	// 500xx: Internal Server Error
	ErrInternalServerErrorGeneral = ErrResponse{http.StatusInternalServerError, 50000, tip.InternalServerErrorTip, nil}
)

func (r ErrResponse) MarshalJSON() ([]byte, error) {
	template := `{"code":%d,"message":"%s"`
	part := fmt.Sprintf(template, r.Code, utils.EscapeDoubleQuotes(r.Tip.String()))
	res := []byte(part)

	// add Extra field
	if r.Extra != nil {
		res = append(res, []byte(`,"extra":`)...)
		if val, err := json.Marshal(r.Extra); err != nil {
			return nil, err
		} else {
			res = append(res, val...)
		}
	}
	res = append(res, '}') // end

	return res, nil
}

// set Extra
func (r ErrResponse) With(Extra interface{}) ErrResponse {
	n := ErrResponse{}

	// deep copy
	if err := utils.DeepCopy(&n, &r); err != nil {
		log.Bg().Error("deep copy ErrResponse fail", zap.Error(err))
	}

	n.Extra = Extra
	return n
}

// todo there may be a bug here, when call With before WithArgs
func (r ErrResponse) WithArgs(args ...interface{}) ErrResponse {
	n := ErrResponse{}

	// deep copy
	if err := utils.DeepCopy(&n, &r); err != nil {
		log.Bg().Error("deep copy ErrResponse fail", zap.Error(err))
	}

	for k, v := range r.Tip {
		str := fmt.Sprintf(v, args...)
		// remove map[ in of output
		// todo There may be a bug here
		str = strings.ReplaceAll(str, "map[", "[")
		n.Tip[k] = str
	}
	return n
}
