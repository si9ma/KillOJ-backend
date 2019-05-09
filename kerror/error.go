// error package for killoj
package kerror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	ErrBadRequestGeneral = ErrResponse{http.StatusBadRequest, 40000, tip.BadRequestGeneralTip, nil}
	ErrArgValidateFail   = ErrResponse{http.StatusBadRequest, 40001, tip.ArgValidateFailTip, nil}
	ErrNotExist          = ErrResponse{http.StatusBadRequest, 40002, tip.NotExistTip, nil}
	ErrAlreadyExist      = ErrResponse{http.StatusBadRequest, 40003, tip.AlreadyExistTip, nil}
	// same as ErrAlreadyExist, but tip is different
	ErrUserAlreadyExistInOrg = ErrResponse{http.StatusBadRequest, 40003, tip.UserAlreadyExistInOrgTip, nil}

	// 401xx:
	ErrUnauthorizedGeneral = ErrResponse{http.StatusUnauthorized, 40100, tip.UnauthorizedGeneralTip, nil}
	ErrUserNotExist        = ErrResponse{http.StatusUnauthorized, 40101, tip.UserNotExistTip, nil}

	// 500xx: Internal Server Error
	ErrInternalServerErrorGeneral = ErrResponse{http.StatusInternalServerError, 50000, tip.InternalServerErrorTip, nil}
)

func (r ErrResponse) MarshalJSON() ([]byte, error) {
	template := `{"code":%d,"message":"%s"`
	part := fmt.Sprintf(template, r.Code, r.Tip.String())
	res := []byte(part)

	// add Extra field
	if r.Extra != nil {
		res = append(res, []byte(`,"Extra":`)...)
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

func (r ErrResponse) WithArgs(args ...interface{}) ErrResponse {
	n := ErrResponse{}

	// deep copy
	if err := utils.DeepCopy(&n, &r); err != nil {
		log.Bg().Error("deep copy ErrResponse fail", zap.Error(err))
	}

	for k, v := range r.Tip {
		n.Tip[k] = fmt.Sprintf(v, args...)
	}
	return n
}
