package filter_test

import (
	"fmt"

	"github.com/ekristen/libnuke/pkg/types"
)

type TestResource struct {
	Props types.Properties
}

func (t *TestResource) GetProperty(key string) (string, error) {
	if key == "no_stringer" { //nolint:staticcheck
		return "", fmt.Errorf("does not support legacy IDs")
	} else if key == "no_properties" {
		return "", fmt.Errorf("does not support custom properties")
	}

	return t.Props[key], nil
}

func (t *TestResource) Properties() types.Properties {
	return t.Props
}
