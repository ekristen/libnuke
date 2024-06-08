package errors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
)

func TestErrorIs(t *testing.T) {
	err := liberrors.ErrSkipRequest("resource is regional")
	var testErr liberrors.ErrSkipRequest
	if !errors.As(err, &testErr) {
		t.Errorf("errors.Is failed")
	}
}

const testStringValue = "this is just a test"

func TestErrors(t *testing.T) {
	cases := []struct {
		err error
	}{
		{
			err: liberrors.ErrSkipRequest(testStringValue),
		},
		{liberrors.ErrUnknownEndpoint(testStringValue)},
		{liberrors.ErrWaitResource(testStringValue)},
		{liberrors.ErrHoldResource(testStringValue)},
		{liberrors.ErrUnknownPreset(testStringValue)},
		{liberrors.ErrDeprecatedResourceType(testStringValue)},
	}

	for _, c := range cases {
		if c.err == nil {
			t.Errorf("error is nil")
		}

		assert.Equal(t, c.err.Error(), testStringValue)
	}
}
