package kerror

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseErr_MarshalJSON(t *testing.T) {
	if res, err := json.Marshal(ArgValidateFail); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
		assert.Equal(t, `{"code":40001,"message":"validate argument fail"}`, string(res))
	}
}

func TestResponseErr_With(t *testing.T) {
	if res, err := json.Marshal(ArgValidateFail.With([]string{"a", "b", "c"})); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
		assert.Equal(t, `{"code":40001,"message":"validate argument fail","extra":["a","b","c"]}`, string(res))
	}
}
