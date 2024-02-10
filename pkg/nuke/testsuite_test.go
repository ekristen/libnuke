package nuke

import (
	"context"
	"fmt"
	"github.com/ekristen/libnuke/pkg/registry"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"
)

type TestGlobalHook struct {
	t  *testing.T
	tf func(t *testing.T, e *logrus.Entry)
}

func (h *TestGlobalHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TestGlobalHook) Fire(e *logrus.Entry) error {
	if h.tf != nil {
		h.tf(h.t, e)
	}

	return nil
}

var (
	TestResourceType         = "testResourceType"
	TestResourceRegistration = &registry.Registration{
		Name:   TestResourceType,
		Scope:  "account",
		Lister: &TestResourceLister{},
	}

	TestResourceType2         = "testResourceType2"
	TestResourceRegistration2 = &registry.Registration{
		Name:   TestResourceType2,
		Scope:  "account",
		Lister: &TestResourceLister{},
		DependsOn: []string{
			TestResourceType,
		},
	}
)

type TestOpts struct {
	Test               *testing.T
	SessionOne         string
	SessionTwo         string
	ThrowError         bool
	ThrowSkipError     bool
	ThrowEndpointError bool
	Panic              bool
	SecondResource     bool
}

type TestResourceLister struct {
	Filtered    bool
	RemoveError bool
}

func (l TestResourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(TestOpts)

	if opts.ThrowError {
		return nil, assert.AnError
	}

	if opts.ThrowSkipError {
		return nil, errors.ErrSkipRequest("skip request error for testing")
	}

	if opts.ThrowEndpointError {
		return nil, errors.ErrUnknownEndpoint("unknown endpoint error for testing")
	}

	if opts.Panic {
		panic(fmt.Errorf("panic error for testing"))
	}

	if opts.SecondResource {
		return []resource.Resource{
			&TestResource2{
				Filtered:    l.Filtered,
				RemoveError: l.RemoveError,
			},
		}, nil
	}

	return []resource.Resource{
		&TestResource{
			Filtered:    l.Filtered,
			RemoveError: l.RemoveError,
		},
	}, nil
}

// --------------------------------------------------------------------------

type TestResource struct {
	Filtered    bool
	RemoveError bool
}

func (r *TestResource) Filter() error {
	if r.Filtered {
		return fmt.Errorf("cannot remove default")
	}

	return nil
}

func (r *TestResource) Remove(_ context.Context) error {
	if r.RemoveError {
		return fmt.Errorf("remove error")
	}
	return nil
}

func (r *TestResource) Settings(setting *settings.Setting) {

}

type TestResource2 struct {
	Filtered    bool
	RemoveError bool
}

func (r *TestResource2) Filter() error {
	if r.Filtered {
		return fmt.Errorf("cannot remove default")
	}

	return nil
}

func (r *TestResource2) Remove(_ context.Context) error {
	if r.RemoveError {
		return fmt.Errorf("remove error")
	}
	return nil
}

func (r *TestResource2) Properties() types.Properties {
	props := types.NewProperties()
	props.Set("test", "testing")
	return props
}
