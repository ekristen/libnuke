package scanner

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/registry"
)

func Test_NewScannerWithMorphOpts(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne: "testing",
	}

	morphOpts := func(o interface{}, resourceType string) interface{} {
		o1 := o.(TestOpts)
		o1.SessionTwo = o1.SessionOne + "-" + resourceType
		return o1
	}

	scanner := New("Owner", []string{testResourceType}, opts)
	mutateErr := scanner.RegisterMutateOptsFunc(morphOpts)
	assert.NoError(t, mutateErr)

	scanner.SetParallelQueries(8)

	err := scanner.Run(context.TODO())
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 1)

	for item := range scanner.Items {
		assert.Equal(t, "testing", item.Opts.(TestOpts).SessionOne)
		assert.Equal(t, "testing-testResourceType", item.Opts.(TestOpts).SessionTwo)
		assert.Equal(t, "OwnerModded", item.Owner)
	}
}

func Test_NewScannerWithDuplicateMorphOpts(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne: "testing",
	}

	morphOpts := func(o interface{}, resourceType string) interface{} {
		o1 := o.(TestOpts)
		o1.SessionTwo = o1.SessionOne + "-" + resourceType
		return o1
	}

	scanner := New("Owner", []string{testResourceType}, opts)
	optErr := scanner.RegisterMutateOptsFunc(morphOpts)
	assert.NoError(t, optErr)

	optErr = scanner.RegisterMutateOptsFunc(morphOpts)
	assert.Error(t, optErr)
}

func Test_NewScannerWithResourceListerError(t *testing.T) {
	registry.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/registry/registry.go") {
				return
			}

			assert.Equal(t, "Listing testResourceType failed:\n    assert.AnError general error for testing", e.Message)
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	registry.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne: "testing",
		ThrowError: true,
	}

	scanner := New("Owner", []string{testResourceType}, opts)
	err := scanner.Run(context.TODO())
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 0)
}

func Test_NewScannerWithInvalidResourceListerError(t *testing.T) {
	registry.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/registry/registry.go") {
				return
			}

			assert.Equal(t, "lister for resource type not found: does-not-exist", e.Message)
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	registry.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne: "testing",
		ThrowError: true,
	}

	scanner := New("Owner", []string{"does-not-exist"}, opts)
	err := scanner.Run(context.TODO())
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 0)
}

func Test_NewScannerWithResourceListerErrorSkip(t *testing.T) {
	registry.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/registry/registry.go") {
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

	registry.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne:     "testing",
		ThrowSkipError: true,
	}

	scanner := New("Owner", []string{testResourceType}, opts)
	err := scanner.Run(context.TODO())
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 0)
}

func Test_NewScannerWithResourceListerErrorUnknownEndpoint(t *testing.T) {
	registry.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/registry/registry.go") {
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

	registry.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne:         "testing",
		ThrowEndpointError: true,
	}

	scanner := New("Owner", []string{testResourceType}, opts)
	err := scanner.Run(context.TODO())
	assert.NoError(t, err)

	assert.Len(t, scanner.Items, 0)
}

func TestRunSemaphoreFirstAcquireError(t *testing.T) {
	// Create a new scanner
	scanner := New("owner", []string{testResourceType}, nil)
	scanner.SetParallelQueries(0)

	// Create a context that will be canceled immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Run the scanner
	err := scanner.Run(ctx)
	assert.Error(t, err)
}

func TestRunSemaphoreSecondAcquireError(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(testResourceRegistration)
	// Create a new scanner
	scanner := New("owner", []string{testResourceType}, TestOpts{
		Sleep: 45 * time.Second,
	})

	// Create a context that will be canceled immediately
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Run the scanner
	err := scanner.Run(ctx)
	assert.Error(t, err)
}

func Test_NewScannerWithResourceListerPanic(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	panicCaught := false

	registry.ClearRegistry()
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/registry/registry.go") {
				assert.Equal(t, "registered resource lister", e.Message)
				wg.Done()
				return
			}

			if strings.HasSuffix(e.Caller.File, "pkg/scanner/scanner.go") && e.Caller.Line == 110 {
				assert.Contains(t, e.Message, "Listing testResourceType failed")
				assert.Contains(t, e.Message, "panic error for testing")
				logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
				panicCaught = true
				wg.Done()
				return
			}
		},
	})

	registry.Register(testResourceRegistration)

	opts := TestOpts{
		SessionOne: "testing",
		Panic:      true,
	}

	scanner := New("Owner", []string{testResourceType}, opts)
	_ = scanner.Run(context.TODO())

	if waitTimeout(&wg, 10*time.Second) {
		t.Fatal("Wait group timed out")
		return
	}

	assert.True(t, panicCaught)
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
