package errors_test

import (
	"errors"
	"testing"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
)

func TestErrorIs(t *testing.T) {
	err := liberrors.ErrSkipRequest("resource is regional")
	var testErr liberrors.ErrSkipRequest
	if !errors.As(err, &testErr) {
		t.Errorf("errors.Is failed")
	}
}
