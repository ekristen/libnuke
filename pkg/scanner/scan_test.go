package scanner

import (
	"context"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/resource"
)

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

	scanner := NewScanner("Owner", []string{testResourceType}, opts)
	mutateErr := scanner.RegisterMutateOptsFunc(morphOpts)
	assert.NoError(t, mutateErr)

	scanner.SetParallelQueries(8)

	err := scanner.Run(context.TODO())
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 1)

	for item := range scanner.Items {
		assert.Equal(t, "testing", item.Opts.(TestOpts).SessionOne)
		assert.Equal(t, "testing-testResourceType", item.Opts.(TestOpts).SessionTwo)
	}
}

func Test_NewScannerWithDuplicateMorphOpts(t *testing.T) {
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

	scanner := NewScanner("Owner", []string{testResourceType}, opts)
	optErr := scanner.RegisterMutateOptsFunc(morphOpts)
	assert.NoError(t, optErr)

	optErr = scanner.RegisterMutateOptsFunc(morphOpts)
	assert.Error(t, optErr)
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

	scanner := NewScanner("Owner", []string{testResourceType}, opts)
	err := scanner.Run(context.TODO())
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

	scanner := NewScanner("Owner", []string{testResourceType}, opts)
	err := scanner.Run(context.TODO())
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

	scanner := NewScanner("Owner", []string{testResourceType}, opts)
	err := scanner.Run(context.TODO())
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

	scanner := NewScanner("Owner", []string{testResourceType}, opts, nil)
	scanner.Run()

	assert.Len(t, scanner.Items, 0)
}
*/
