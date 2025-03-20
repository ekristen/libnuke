package nuke

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	liberrors "github.com/ekristen/libnuke/pkg/errors"

	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/scanner"
	"github.com/ekristen/libnuke/pkg/settings"
)

var testParameters = &Parameters{
	Force:      true,
	ForceSleep: 3,
	Quiet:      false,
	NoDryRun:   false,
}

var testParametersRemove = &Parameters{
	Force:      true,
	ForceSleep: 3,
	Quiet:      true,
	NoDryRun:   true,
}

var testParametersGroups = &Parameters{
	Force:           true,
	ForceSleep:      3,
	Quiet:           false,
	NoDryRun:        false,
	UseFilterGroups: true,
}

const testScope registry.Scope = "test"

func Test_Nuke_Version(t *testing.T) {
	logger := logrus.WithField("test", true)

	assertions := 0

	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if !strings.HasSuffix(e.Caller.File, "pkg/nuke/nuke.go") {
				return
			}

			if e.Caller.Line == 351 {
				assert.Equal(t, "1.0.0-test", e.Message)
				assertions++
			}
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(nil)

	n := New(testParameters, nil, nil)
	n.SetLogger(logger)
	n.SetRunSleep(time.Millisecond * 5)

	n.RegisterVersion("1.0.0-test")

	// Call the Version function
	n.Version()

	assert.Equal(t, 1, assertions)
}

func TestNuke_Settings(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)
	n.Settings = &settings.Settings{
		"TestResource": &settings.Setting{
			"DisableDeletionProtection": true,
		},
	}

	testResourceSettings := n.Settings.Get("TestResource")
	assert.NotNil(t, testResourceSettings)
	assert.Equal(t, true, testResourceSettings.Get("DisableDeletionProtection"))
}

func Test_Nuke_Validators_Default(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	err := n.Validate()
	assert.NoError(t, err)
}

func Test_Nuke_Validators_Register1(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	n.RegisterValidateHandler(func() error {
		return fmt.Errorf("validator called")
	})

	err := n.Validate()
	assert.Error(t, err)
	assert.Equal(t, "validator called", err.Error())
}

func Test_Nuke_Validators_Register2(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	n.RegisterValidateHandler(func() error {
		return fmt.Errorf("validator called")
	})

	n.RegisterValidateHandler(func() error {
		return fmt.Errorf("second validator called")
	})

	assert.Len(t, n.ValidateHandlers, 2)
}

func Test_Nuke_Validators_Error(t *testing.T) {
	p := &Parameters{
		Force:      true,
		ForceSleep: 1,
		Quiet:      true,
	}
	n := New(p, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	err := n.Validate()
	assert.Error(t, err)
	assert.Equal(t, "value for --force-sleep cannot be less than 3 seconds. This is for your own protection", err.Error())
}

func Test_Nuke_ResourceTypes(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	n.RegisterResourceTypes(testScope, "TestResource")

	assert.Len(t, n.ResourceTypes[testScope], 1)
}

func Test_Nuke_Scanners(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := struct {
		name string
	}{
		name: "test",
	}

	s, err := scanner.New(&scanner.Config{
		Owner:         "test",
		ResourceTypes: []string{"TestResource"},
		Opts:          opts,
	})
	assert.NoError(t, err)

	err = n.RegisterScanner(testScope, s)
	assert.NoError(t, err)

	assert.Len(t, n.Scanners[testScope], 1)
}

func Test_Nuke_Scanners_Duplicate(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := struct {
		name string
	}{
		name: "test",
	}

	s, err := scanner.New(&scanner.Config{
		Owner:         "test",
		ResourceTypes: []string{"TestResource"},
		Opts:          opts,
	})
	assert.NoError(t, err)

	err = n.RegisterScanner(testScope, s)
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, s)
	assert.Error(t, sErr)

	assert.Len(t, n.Scanners[testScope], 1)
}

func TestNuke_RegisterMultipleScanners(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := struct {
		name string
	}{
		name: "test",
	}

	var mutateOpts = func(o interface{}, resourceType string) interface{} {
		return o
	}

	s, err := scanner.New(&scanner.Config{
		Owner:         "test",
		ResourceTypes: []string{"TestResource"},
		Opts:          opts,
	})
	assert.NoError(t, err)
	assert.NoError(t, s.RegisterMutateOptsFunc(mutateOpts))

	s2, err := scanner.New(&scanner.Config{
		Owner:         "test2",
		ResourceTypes: []string{"TestResource"},
		Opts:          opts,
	})
	assert.NoError(t, err)
	assert.NoError(t, s2.RegisterMutateOptsFunc(mutateOpts))

	assert.NoError(t, n.RegisterScanner(testScope, s))
	assert.NoError(t, n.RegisterScanner(testScope, s2))
	assert.Len(t, n.Scanners[testScope], 2)
}

func Test_Nuke_RegisterPrompt(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	n.RegisterPrompt(func() error {
		return fmt.Errorf("prompt error")
	})

	err := n.Prompt()
	assert.Error(t, err)
	assert.Equal(t, "prompt error", err.Error())
}

// ------------------------------------------------------

func Test_Nuke_Scan(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration)
	registry.Register(&registry.Registration{
		Name:  TestResourceType2,
		Scope: "account",
		Lister: TestResourceLister{
			Filtered: true,
		},
	})

	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne: "testing",
	}
	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType, TestResourceType2},
		Opts:          opts,
	})
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Scan(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, 2, n.Queue.Total())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateNew))
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFiltered))
}

// ---------------------------------------------------------------------

type TestResource3 struct {
	Error bool
}

func (r *TestResource3) Remove(_ context.Context) error {
	if r.Error {
		return fmt.Errorf("remove error")
	}
	return nil
}

func Test_Nuke_HandleRemove(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	i := &queue.Item{
		Resource: &TestResource3{},
		State:    queue.ItemStateNew,
	}

	n.HandleRemove(context.TODO(), i)
	assert.Equal(t, queue.ItemStatePending, i.State)
}

func Test_Nuke_HandleRemoveError(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	i := &queue.Item{
		Resource: &TestResource3{
			Error: true,
		},
		State: queue.ItemStateNew,
	}

	n.HandleRemove(context.TODO(), i)
	assert.Equal(t, queue.ItemStateFailed, i.State)
}

// ------------------------------------------------------------

func Test_Nuke_Run(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration)

	p := &Parameters{
		Force:      true,
		ForceSleep: 3,
		Quiet:      true,
		NoDryRun:   true,
	}

	n := New(p, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne: "testing",
	}
	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType},
		Opts:          opts,
	})
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Run(context.TODO())
	assert.NoError(t, err)
}

func Test_Nuke_Run_Error(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:  TestResourceType2,
		Scope: "account",
		Lister: TestResourceLister{
			RemoveError: true,
		},
	})

	p := &Parameters{
		Force:      true,
		ForceSleep: 3,
		Quiet:      true,
		NoDryRun:   true,
	}
	n := New(p, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne: "testing",
	}
	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType2},
		Opts:          opts,
	})

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Run(context.TODO())
	assert.NoError(t, err)
}

// ------------------------------------------------------------

var TestResource4Resources []resource.Resource

type TestResource4 struct {
	id       string
	parentID string
}

func (r *TestResource4) Remove(_ context.Context) error {
	if r.parentID != "" {
		parentFound := false

		for _, o := range TestResource4Resources {
			id := o.(resource.LegacyStringer).String()
			if id == r.parentID {
				parentFound = true
			}
		}

		if parentFound {
			return liberrors.ErrHoldResource("waiting for parent to be removed")
		}
	}

	return nil
}

func (r *TestResource4) String() string {
	return r.id
}

type TestResource4Lister struct {
	attempts int
}

func (l *TestResource4Lister) List(_ context.Context, _ interface{}) ([]resource.Resource, error) {
	l.attempts++

	if l.attempts == 1 {
		for x := 0; x < 5; x++ {
			if x == 0 {
				TestResource4Resources = append(TestResource4Resources, &TestResource4{
					id:       fmt.Sprintf("resource-%d", x),
					parentID: "",
				})
			} else {
				TestResource4Resources = append(TestResource4Resources, &TestResource4{
					id:       fmt.Sprintf("resource-%d", x),
					parentID: "resource-0",
				})
			}
		}
	} else if l.attempts > 3 {
		TestResource4Resources = TestResource4Resources[1:]
	}

	return TestResource4Resources, nil
}

func Test_Nuke_Run_ItemStateHold(t *testing.T) {
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResource4",
		Scope:  testScope,
		Lister: &TestResource4Lister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResource4"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)
	assert.Equal(t, 5, n.Queue.Count(queue.ItemStateFinished))
}
