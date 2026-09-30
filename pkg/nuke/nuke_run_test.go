package nuke

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/scanner"
	"github.com/ekristen/libnuke/pkg/settings"
)

type TestResourceSuccess struct {
	id       string
	Filtered bool
}

func (r *TestResourceSuccess) ID() string {
	if r.id == "" {
		// Note: math/rand is acceptable for this unit test, it does not need to be crypto/rand
		r3 := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
		r.id = fmt.Sprintf("TestResourceSuccess-%d", r3.Intn(100))
	}
	return r.id
}

func (r *TestResourceSuccess) Remove(_ context.Context) error { return nil }
func (r *TestResourceSuccess) Filter() error {
	if r.Filtered {
		return fmt.Errorf("filtered")
	}
	return nil
}
func (r *TestResourceSuccess) String() string {
	return r.id
}

type TestResourceSuccessLister struct {
	listed bool
}

func (l *TestResourceSuccessLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	if l.listed {
		return []resource.Resource{}, nil
	}
	l.listed = true
	return []resource.Resource{&TestResourceSuccess{}, &TestResourceSuccess{Filtered: true}}, nil
}

type TestResourceFailure struct{}

func (r *TestResourceFailure) Remove(_ context.Context) error {
	return fmt.Errorf("unable to remove")
}
func (r *TestResourceFailure) String() string { return "TestResourceFailure" }

type TestResourceFailureLister struct{}

func (l *TestResourceFailureLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	return []resource.Resource{&TestResourceFailure{}}, nil
}

type TestResourceWaitLister struct{}

func (l *TestResourceWaitLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	return []resource.Resource{&TestResourceSuccess{}}, nil
}

// Test_Nuke_Run_Simple tests a simple run with no dry run enabled so all resources are removed.
func Test_Nuke_Run_Simple(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceSuccess"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateNew))
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFiltered))
	assert.Equal(t, 2, n.Queue.Total())
}

// Test_Nuke_Run_ScanError tests a simple run with no dry run enabled so all resources are removed.
func Test_Nuke_Run_ScanError(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceSuccess"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(ctx)
	assert.Error(t, runErr)
}

// Test_NukeRunSimpleWithFirstPromptError tests the first prompt throwing an error
func Test_NukeRunSimpleWithFirstPromptError(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)
	n.RegisterPrompt(func() error {
		return fmt.Errorf("first prompt called")
	})

	runErr := n.Run(context.TODO())
	assert.Error(t, runErr)
	assert.Equal(t, "first prompt called", runErr.Error())
}

// Test_NukeRunSimpleWithFirstPromptError tests the second prompt throwing an error
func Test_NukeRunSimpleWithSecondPromptError(t *testing.T) {
	promptCalled := false
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)
	n.RegisterPrompt(func() error {
		if promptCalled {
			return fmt.Errorf("second prompt called")
		}

		promptCalled = true

		return nil
	})

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceSuccess"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.Error(t, runErr)
	assert.Equal(t, "second prompt called", runErr.Error())
}

// Test_Nuke_Run_SimpleWithNoDryRun tests a simple run with no dry run enabled so all resources are removed.
func Test_Nuke_Run_SimpleWithNoDryRun(t *testing.T) {
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceSuccess"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFinished))
}

// Test_Nuke_Run_Failure tests a run with a resource that fails to remove, so it should be in the failed state.
// It also tests that a resource is successfully removed as well, to test the entire fail state.
func Test_Nuke_Run_Failure(t *testing.T) {
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	registry.Register(&registry.Registration{
		Name:   "TestResourceFailure",
		Lister: &TestResourceFailureLister{},
	})

	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceSuccess", "TestResourceFailure"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.Error(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFinished))
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFailed))
	assert.Equal(t, testParametersRemove.MaxFailureRetries, n.failedCount)
}

var testParametersMaxWaitRetries = &Parameters{
	Force:          true,
	ForceSleep:     3,
	Quiet:          true,
	NoDryRun:       true,
	MaxWaitRetries: 3,
}

func Test_NukeRunWithMaxWaitRetries(t *testing.T) {
	n := New(testParametersMaxWaitRetries, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceWaitLister{},
	})

	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceSuccess"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.Error(t, runErr)
	assert.Equal(t, "max wait retries of 3 exceeded", runErr.Error())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateWaiting))
}

// ---------------------

var testResourceAlphaFilterSecondFail = false

type TestResourceAlpha struct {
	WaitCount   int
	WaitFail    bool
	FilterCount int
	FilterFail  bool
}

func (r *TestResourceAlpha) Remove(_ context.Context) error { return nil }
func (r *TestResourceAlpha) String() string                 { return "TestResourceAlpha" }
func (r *TestResourceAlpha) Filter() error {
	if !r.FilterFail {
		return nil
	}

	if testResourceAlphaFilterSecondFail {
		return fmt.Errorf("filter operation failed")
	}

	testResourceAlphaFilterSecondFail = true

	return nil
}
func (r *TestResourceAlpha) Settings(_ *settings.Setting) {}
func (r *TestResourceAlpha) HandleWait(_ context.Context) error {
	r.WaitCount++

	if r.WaitCount < 3 {
		return errors.ErrWaitResource("still waiting")
	}

	if r.WaitFail {
		return fmt.Errorf("wait operation failed")
	}

	return nil
}

type TestResourceAlphaLister struct {
	listed         bool
	waitFail       bool
	failListOnWait bool
}

func (l *TestResourceAlphaLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	if l.listed && l.failListOnWait {
		return nil, fmt.Errorf("unable to list")
	}
	if l.listed {
		return []resource.Resource{}, nil
	}
	l.listed = true
	return []resource.Resource{&TestResourceAlpha{
		WaitFail: l.waitFail,
	}}, nil
}

func TestNuke_RunWithWaitOnDependencies(t *testing.T) {
	n := New(&Parameters{
		Force:              true,
		ForceSleep:         3,
		Quiet:              true,
		NoDryRun:           true,
		WaitOnDependencies: true,
	}, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResourceAlpha",
		Lister: &TestResourceAlphaLister{},
	})
	registry.Register(&registry.Registration{
		Name:   "TestResourceBeta",
		Lister: &TestResourceAlphaLister{},
		DependsOn: []string{
			"TestResourceAlpha",
		},
	})

	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceAlpha", "TestResourceBeta"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 2, n.Queue.Count(queue.ItemStateFinished))
}

func TestNuke_RunWithHandleWaitFail(t *testing.T) {
	n := New(&Parameters{
		Force:              true,
		ForceSleep:         3,
		Quiet:              true,
		NoDryRun:           true,
		WaitOnDependencies: true,
	}, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name: "TestResourceAlpha",
		Lister: &TestResourceAlphaLister{
			waitFail: true,
		},
	})

	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceAlpha"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	if assert.Error(t, runErr) {
		assert.Equal(t, "failed", runErr.Error())
	}

	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFinished))
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFailed))
}

func TestNuke_RunNoResources(t *testing.T) {
	n := New(&Parameters{
		Force:              true,
		ForceSleep:         3,
		Quiet:              true,
		NoDryRun:           true,
		WaitOnDependencies: true,
	}, nil, nil)
	n.SetLogger(logrus.WithField("test", true))

	registry.ClearRegistry()

	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateNew))
}

func TestNuke_HandleWaitListError(t *testing.T) {
	n := New(&Parameters{
		Force:              true,
		ForceSleep:         3,
		Quiet:              true,
		NoDryRun:           true,
		WaitOnDependencies: true,
	}, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name: "TestResourceAlpha",
		Lister: &TestResourceAlphaLister{
			failListOnWait: true,
		},
	})

	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceAlpha"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	if assert.Error(t, runErr) {
		assert.Equal(t, "failed", runErr.Error())
	}

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFailed))
}

type TestResourceBetaLister struct {
}

func (l *TestResourceBetaLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	return []resource.Resource{&TestResourceAlpha{
		FilterFail: true,
	}}, nil
}

func TestNuke_SecondFilterFail(t *testing.T) {
	n := New(&Parameters{
		Force:              true,
		ForceSleep:         3,
		Quiet:              true,
		NoDryRun:           true,
		WaitOnDependencies: true,
	}, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResourceBeta",
		Lister: &TestResourceBetaLister{},
	})

	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestResourceBeta"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)
}

// ---------------------

// TestResourceStuck is accepted for deletion but then stays listed in a state it never leaves
type TestResourceStuck struct {
	removed   *bool
	state     string
	filterErr func(string) error
}

func (r *TestResourceStuck) Remove(_ context.Context) error {
	*r.removed = true
	return nil
}
func (r *TestResourceStuck) String() string { return "TestResourceStuck" }
func (r *TestResourceStuck) Filter() error {
	if r.state == "ACTIVE" {
		return nil
	}
	return r.filterErr(r.state)
}

type TestResourceStuckLister struct {
	removed   bool
	filterErr func(string) error
}

func (l *TestResourceStuckLister) List(_ context.Context, _ interface{}) ([]resource.Resource, error) {
	state := "ACTIVE"
	if l.removed {
		state = "FAILED"
	}
	return []resource.Resource{&TestResourceStuck{removed: &l.removed, state: state, filterErr: l.filterErr}}, nil
}

// TestNuke_RunWithStuckResource tests that a resource still listed after removal only counts as finished when its
// filter does not report it as failed
func TestNuke_RunWithStuckResource(t *testing.T) {
	cases := []struct {
		name      string
		filterErr func(string) error
		wantErr   bool
		finished  int
		failed    int
	}{
		{
			name:      "filtered-counts-as-finished",
			filterErr: func(state string) error { return fmt.Errorf("resource is %s", state) },
			finished:  1,
		},
		{
			name:      "failed-resource-fails-run",
			filterErr: func(state string) error { return errors.ErrFailedResource("resource is " + state) },
			wantErr:   true,
			failed:    1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := New(testParametersRemove, nil, nil)
			n.SetLogger(logrus.WithField("test", true))
			n.SetRunSleep(time.Millisecond * 5)

			registry.ClearRegistry()
			registry.Register(&registry.Registration{
				Name:   "TestResourceStuck",
				Lister: &TestResourceStuckLister{filterErr: tc.filterErr},
			})

			newScanner, err := scanner.New(&scanner.Config{
				Owner:         "Owner",
				ResourceTypes: []string{"TestResourceStuck"},
				Opts:          nil,
			})
			assert.NoError(t, err)

			scannerErr := n.RegisterScanner(testScope, newScanner)
			assert.NoError(t, scannerErr)

			runErr := n.Run(context.TODO())
			if tc.wantErr {
				assert.Error(t, runErr)
			} else {
				assert.NoError(t, runErr)
			}

			assert.Equal(t, tc.finished, n.Queue.Count(queue.ItemStateFinished))
			assert.Equal(t, tc.failed, n.Queue.Count(queue.ItemStateFailed))
		})
	}
}
