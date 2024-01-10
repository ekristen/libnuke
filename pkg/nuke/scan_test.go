package nuke

import (
	"flag"
	"fmt"
	"github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
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

type TestResource struct{}

func (r TestResource) Remove() error {
	return nil
}

type TestResourceLister struct{}

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

	return []resource.Resource{
		TestResource{},
	}, nil
}

type TestOpts struct {
	SessionOne         string
	SessionTwo         string
	ThrowError         bool
	ThrowSkipError     bool
	ThrowEndpointError bool
	Panic              bool
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

func Test_NewScannerWithDependsOn(t *testing.T) {
	resource.ClearRegistry()

	resource.Register(testResourceRegistration)
	resource.Register(testResourceRegistration2)

	opts := TestOpts{
		SessionOne: "testing",
	}

	scanner := NewScanner("owner", []string{testResourceType, testResourceType2}, opts)

	err := scanner.Run()
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 2)

	x := 0
	for item := range scanner.Items {
		switch item.Type {
		case testResourceType:
			assert.Equal(t, testResourceType, item.Type)
			assert.Equal(t, queue.ItemStateNew, item.State)
		case testResourceType2:
			assert.Equal(t, testResourceType2, item.Type)
			assert.Equal(t, queue.ItemStateNewDependency, item.State)
		}
		x++
	}
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
				assert.Equal(t, logrus.WarnLevel, e.Level)
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
