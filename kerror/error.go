// error package for killoj
package kerror

import (
	"encoding/json"
	"fmt"

	"github.com/si9ma/KillOJ-common/tip"
)

type ResponseErr struct {
	code  int
	tip   tip.Tip
	extra interface{}
}

var (
	// 400xx : bad request
	BadRequestGeneral = ResponseErr{40000, nil, nil}
	ArgValidateFail   = ResponseErr{40001, tip.ArgValidateFailTip, nil}
)

func (r ResponseErr) MarshalJSON() ([]byte, error) {
	template := `{"code":%d,"message":"%s"`
	part := fmt.Sprintf(template, r.code, r.tip.String())
	res := []byte(part)

	// add extra field
	if r.extra != nil {
		res = append(res, []byte(`,"extra":`)...)
		if val, err := json.Marshal(r.extra); err != nil {
			return nil, err
		} else {
			res = append(res, val...)
		}
	}
	res = append(res, '}') // end

	return res, nil
}

// set extra
func (r ResponseErr) With(extra interface{}) ResponseErr {
	res := r
	res.extra = extra
	return res
}
