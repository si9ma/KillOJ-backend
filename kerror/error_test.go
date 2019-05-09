package kerror

import (
	"testing"

	"github.com/ulule/deepcopier"

	"github.com/si9ma/KillOJ-common/tip"

	"github.com/stretchr/testify/assert"
)

func TestErrResponse_WithArgs(t *testing.T) {
	assert.Equal(t, "test not exist", ErrNotExist.WithArgs("test").Tip.String())
}

func TestDeepCopy(t *testing.T) {
	ti := tip.UserAlreadyExistInOrgTip
	tii := tip.Tip{}
	if err := deepcopier.Copy(&ti).To(&tii); err != nil {
		t.Fatal(err)
	}
}
