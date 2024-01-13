package nuke

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/featureflag"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

func init() {
	if flag.Lookup("test.v") != nil {
		logrus.SetOutput(io.Discard)
	}
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)
}

var (
	testResourceType         = "testResourceType"
	testResourceRegistration = resource.Registration{
		Name:   testResourceType,
		Scope:  "account",
		Lister: TestResourceLister{},
	}

	testResourceType2         = "testResourceType2"
	testResourceRegistration2 = resource.Registration{
		Name:   testResourceType2,
		Scope:  "account",
		Lister: TestResourceLister{},
		DependsOn: []string{
			testResourceType,
		},
	}
)

type TestResource struct {
	Filtered    bool
	RemoveError bool
}

func (r TestResource) Filter() error {
	if r.Filtered {
		return fmt.Errorf("cannot remove default")
	}

	return nil
}

func (r TestResource) Remove() error {
	if r.RemoveError {
		return fmt.Errorf("remove error")
	}
	return nil
}

func (r TestResource) FeatureFlags(ff *featureflag.FeatureFlags) {

}

type TestResource2 struct {
	Filtered    bool
	RemoveError bool
}

func (r TestResource2) Filter() error {
	if r.Filtered {
		return fmt.Errorf("cannot remove default")
	}

	return nil
}

func (r TestResource2) Remove() error {
	if r.RemoveError {
		return fmt.Errorf("remove error")
	}
	return nil
}

func (r TestResource2) FeatureFlags(ff *featureflag.FeatureFlags) {

}

func (r TestResource2) Properties() types.Properties {
	props := types.NewProperties()
	props.Set("test", "testing")
	return props
}

type TestResourceLister struct {
	Filtered    bool
	RemoveError bool
}

func (l TestResourceLister) List(o interface{}) ([]resource.Resource, error) {
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

type TestOpts struct {
	SessionOne         string
	SessionTwo         string
	ThrowError         bool
	ThrowSkipError     bool
	ThrowEndpointError bool
	Panic              bool
	SecondResource     bool
}

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

func Test_NewScannerWithMorphOpts(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne: "testing",
	}

	morphOpts := func(o interface{}, resourceType string) interface{} {
		o1 := o.(TestOpts)
		o1.SessionTwo = o1.SessionOne + "-" + resourceType
		return o1
	}

	scanner := NewScanner("owner", []string{testResourceType}, opts)
	scanner.RegisterMutateOptsFunc(morphOpts)

	err := scanner.Run()
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 1)

	for item := range scanner.Items {
		assert.Equal(t, "testing", item.Opts.(TestOpts).SessionOne)
		assert.Equal(t, "testing-testResourceType", item.Opts.(TestOpts).SessionTwo)
	}
}

func Test_NewScannerWithResourceListerError(t *testing.T) {
	resource.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/resource/registry.go") {
				return
			}

			assert.Equal(t, "Listing testResourceType failed:\n    assert.AnError general error for testing", e.Message)
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	resource.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne: "testing",
		ThrowError: true,
	}

	scanner := NewScanner("owner", []string{testResourceType}, opts)
	err := scanner.Run()
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 0)
}

func Test_NewScannerWithResourceListerErrorSkip(t *testing.T) {
	resource.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/resource/registry.go") {
				assert.Equal(t, logrus.TraceLevel, e.Level)
				assert.Equal(t, "registered resource lister", e.Message)
				return
			}

			if strings.HasSuffix(e.Caller.File, "pkg/nuke/scan.go") {
				assert.Equal(t, logrus.DebugLevel, e.Level)
				assert.Equal(t, "skipping request: skip request error for testing", e.Message)
			}
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	resource.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne:     "testing",
		ThrowSkipError: true,
	}

	scanner := NewScanner("owner", []string{testResourceType}, opts)
	err := scanner.Run()
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 0)
}

func Test_NewScannerWithResourceListerErrorUnknownEndpoint(t *testing.T) {
	resource.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/resource/registry.go") {
				assert.Equal(t, logrus.TraceLevel, e.Level)
				assert.Equal(t, "registered resource lister", e.Message)
				return
			}

			if strings.HasSuffix(e.Caller.File, "pkg/nuke/scan.go") {
				assert.Equal(t, logrus.DebugLevel, e.Level)
				assert.Equal(t, "skipping request: unknown endpoint error for testing", e.Message)
			}
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	resource.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne:         "testing",
		ThrowEndpointError: true,
	}

	scanner := NewScanner("owner", []string{testResourceType}, opts)
	err := scanner.Run()
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 0)
}

/*
TODO: fix - when run as a whole, this panics but doesn't get caught properly instead the test suite panics and exits

func Test_NewScannerWithResourceListerPanic(t *testing.T) {
	resource.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/resource/registry.go") {
				assert.Equal(t, logrus.TraceLevel, e.Level)
				assert.Equal(t, "registered resource lister", e.Message)
				return
			}

			if strings.HasSuffix(e.Caller.File, "pkg/nuke/scan.go") {
				assert.Contains(t, e.Message, "Listing testResourceType failed:\n assert.AnError general error for testing")
				assert.Contains(t, e.Message, "goroutine")
				assert.Contains(t, e.Message, "runtime/debug.Stack()")
				logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
			}
		},
	})

	resource.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne: "testing",
		Panic:      true,
	}

	scanner := NewScanner("owner", []string{testResourceType}, opts, nil)
	scanner.Run()

	assert.Len(t, scanner.Items, 0)
}
*/
