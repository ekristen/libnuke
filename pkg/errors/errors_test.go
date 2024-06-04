package errors_test

import (
	"errors"
	error2 "github.com/ekristen/libnuke/pkg/errors"
	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"testing"
)

func TestErrorIs(t *testing.T) {
	err := error2.ErrSkipRequest("resource is regional")
	var testErr liberrors.ErrSkipRequest
	if !errors.As(err, &testErr) {
		t.Errorf("errors.Is failed")
	}
}
