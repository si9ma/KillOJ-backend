// error package for killoj
package kerror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	BadRequestGeneral = ErrResponse{http.StatusBadRequest, 40000, tip.BadRequestGeneralTip, nil}
	ArgValidateFail   = ErrResponse{http.StatusBadRequest, 40001, tip.ArgValidateFailTip, nil}

	// 500xx: Internal Server Error
	InternalServerErrorGeneral = ErrResponse{http.StatusInternalServerError, 50000, tip.InternalServerErrorTip, nil}
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
	res := r
	res.Extra = Extra
	return res
}
