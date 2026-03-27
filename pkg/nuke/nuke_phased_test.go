package nuke

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/scanner"
)

// --- Test phased resources ---

// TestPhasedResource implements a resource with two sync phases (disable protection + delete)
type TestPhasedResource struct {
	DisableCalled bool
	DeleteCalled  bool
}

func (r *TestPhasedResource) Remove(_ context.Context) error { return nil }
func (r *TestPhasedResource) String() string                 { return "TestPhasedResource" }

func (r *TestPhasedResource) Phases() []resource.Phase {
	return []resource.Phase{
		{Name: "disable-protection", Run: r.disableProtection},
		{Name: "delete", Run: r.delete},
	}
}

func (r *TestPhasedResource) disableProtection(_ context.Context) error {
	r.DisableCalled = true
	return nil
}

func (r *TestPhasedResource) delete(_ context.Context) error {
	r.DeleteCalled = true
	return nil
}

type TestPhasedResourceLister struct {
	listed bool
}

func (l *TestPhasedResourceLister) List(_ context.Context, _ interface{}) ([]resource.Resource, error) {
	if l.listed {
		return []resource.Resource{}, nil
	}
	l.listed = true
	return []resource.Resource{&TestPhasedResource{}}, nil
}

// TestPhasedResourceAsync implements a resource with async phases
type TestPhasedResourceAsync struct {
	disableWaitCount int
	deleteWaitCount  int
}

func (r *TestPhasedResourceAsync) Remove(_ context.Context) error { return nil }
func (r *TestPhasedResourceAsync) String() string                 { return "TestPhasedResourceAsync" }

func (r *TestPhasedResourceAsync) Phases() []resource.Phase {
	return []resource.Phase{
		{Name: "disable-protection", Run: r.disableProtection},
		{Name: "delete", Run: r.delete},
	}
}

func (r *TestPhasedResourceAsync) disableProtection(_ context.Context) error {
	r.disableWaitCount++
	if r.disableWaitCount < 3 {
		return errors.ErrWaitResource("waiting for disable")
	}
	return nil
}

func (r *TestPhasedResourceAsync) delete(_ context.Context) error {
	r.deleteWaitCount++
	if r.deleteWaitCount < 2 {
		return errors.ErrWaitResource("waiting for delete")
	}
	return nil
}

type TestPhasedResourceAsyncLister struct {
	listed bool
}

func (l *TestPhasedResourceAsyncLister) List(_ context.Context, _ interface{}) ([]resource.Resource, error) {
	if l.listed {
		return []resource.Resource{}, nil
	}
	l.listed = true
	return []resource.Resource{&TestPhasedResourceAsync{}}, nil
}

// TestPhasedResourceFail implements a resource where a phase fails
type TestPhasedResourceFail struct {
	attempts int
}

func (r *TestPhasedResourceFail) Remove(_ context.Context) error { return nil }
func (r *TestPhasedResourceFail) String() string                 { return "TestPhasedResourceFail" }

func (r *TestPhasedResourceFail) Phases() []resource.Phase {
	return []resource.Phase{
		{Name: "disable-protection", Run: r.disableProtection},
		{Name: "delete", Run: r.delete},
	}
}

func (r *TestPhasedResourceFail) disableProtection(_ context.Context) error {
	return nil
}

func (r *TestPhasedResourceFail) delete(_ context.Context) error {
	r.attempts++
	return fmt.Errorf("delete failed")
}

type TestPhasedResourceFailLister struct{}

func (l *TestPhasedResourceFailLister) List(_ context.Context, _ interface{}) ([]resource.Resource, error) {
	return []resource.Resource{&TestPhasedResourceFail{}}, nil
}

// TestPhasedResourceReset implements a resource that resets phases
type TestPhasedResourceReset struct {
	resetCount     int
	disableCalls   int
	deleteCalls    int
	maxResets      int
	deleteSucceeds bool
}

func (r *TestPhasedResourceReset) Remove(_ context.Context) error { return nil }
func (r *TestPhasedResourceReset) String() string                 { return "TestPhasedResourceReset" }

func (r *TestPhasedResourceReset) Phases() []resource.Phase {
	return []resource.Phase{
		{Name: "disable-protection", Run: r.disableProtection},
		{Name: "delete", Run: r.delete},
	}
}

func (r *TestPhasedResourceReset) disableProtection(_ context.Context) error {
	r.disableCalls++
	return nil
}

func (r *TestPhasedResourceReset) delete(_ context.Context) error {
	r.deleteCalls++
	if r.deleteSucceeds {
		return nil
	}
	if r.resetCount < r.maxResets {
		r.resetCount++
		return errors.ErrResetPhases("need to start over")
	}
	// After max resets, succeed
	r.deleteSucceeds = true
	return nil
}

type TestPhasedResourceResetLister struct {
	listed bool
}

func (l *TestPhasedResourceResetLister) List(_ context.Context, _ interface{}) ([]resource.Resource, error) {
	if l.listed {
		return []resource.Resource{}, nil
	}
	l.listed = true
	return []resource.Resource{&TestPhasedResourceReset{maxResets: 1}}, nil
}

// --- Tests ---

func TestNuke_PhasedResource_SyncPhases(t *testing.T) {
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestPhasedResource",
		Lister: &TestPhasedResourceLister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestPhasedResource"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFinished))
}

func TestNuke_PhasedResource_AsyncPhases(t *testing.T) {
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestPhasedResourceAsync",
		Lister: &TestPhasedResourceAsyncLister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestPhasedResourceAsync"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFinished))
}

func TestNuke_PhasedResource_FailRetries(t *testing.T) {
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestPhasedResourceFail",
		Lister: &TestPhasedResourceFailLister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestPhasedResourceFail"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.Error(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFailed))
	// Verify it retried at the delete phase (phase index should still be 1, not 0)
	for _, item := range n.Queue.GetItems() {
		if item.GetState() == queue.ItemStateFailed {
			assert.Equal(t, 1, item.PhaseIndex, "should retry at the failed phase, not reset to 0")
		}
	}
}

func TestNuke_PhasedResource_ResetPhases(t *testing.T) {
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestPhasedResourceReset",
		Lister: &TestPhasedResourceResetLister{},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestPhasedResourceReset"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFinished))
}

func TestNuke_PhasedResource_WithDependencies(t *testing.T) {
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
		Name:   "TestPhasedResource",
		Lister: &TestPhasedResourceLister{},
	})
	registry.Register(&registry.Registration{
		Name:   "TestPhasedResourceDependent",
		Lister: &TestPhasedResourceLister{},
		DependsOn: []string{
			"TestPhasedResource",
		},
	})

	s, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{"TestPhasedResource", "TestPhasedResourceDependent"},
		Opts:          nil,
	})
	assert.NoError(t, err)

	scannerErr := n.RegisterScanner(testScope, s)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 2, n.Queue.Count(queue.ItemStateFinished))
}

func TestNuke_HandlePhase_Direct(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))

	res := &TestPhasedResource{}
	item := &queue.Item{
		Resource: res,
		State:    queue.ItemStateNew,
		Type:     "TestPhasedResource",
	}

	// First call: runs disable-protection phase
	n.HandlePhase(context.TODO(), item, res, make(ListCache))
	assert.Equal(t, queue.ItemStatePending, item.State)
	assert.Equal(t, 1, item.PhaseIndex)
	assert.True(t, res.DisableCalled)
	assert.False(t, res.DeleteCalled)

	// Second call: runs delete phase
	n.HandlePhase(context.TODO(), item, res, make(ListCache))
	assert.Equal(t, queue.ItemStatePending, item.State)
	assert.Equal(t, 2, item.PhaseIndex)
	assert.True(t, res.DeleteCalled)
}

func TestNuke_HandlePhase_HoldError(t *testing.T) {
	n := New(testParameters, nil, nil)
	n.SetLogger(logrus.WithField("test", true))

	res := &TestPhasedResourceHold{}
	item := &queue.Item{
		Resource: res,
		State:    queue.ItemStateNew,
		Type:     "TestPhasedResourceHold",
	}

	n.HandlePhase(context.TODO(), item, res, make(ListCache))
	assert.Equal(t, queue.ItemStateHold, item.State)
}

// TestPhasedResourceHold is a phased resource where the first phase returns ErrHoldResource
type TestPhasedResourceHold struct{}

func (r *TestPhasedResourceHold) Remove(_ context.Context) error { return nil }
func (r *TestPhasedResourceHold) String() string                 { return "TestPhasedResourceHold" }

func (r *TestPhasedResourceHold) Phases() []resource.Phase {
	return []resource.Phase{
		{Name: "wait-for-parent", Run: r.waitForParent},
		{Name: "delete", Run: r.delete},
	}
}

func (r *TestPhasedResourceHold) waitForParent(_ context.Context) error {
	return errors.ErrHoldResource("parent not ready")
}

func (r *TestPhasedResourceHold) delete(_ context.Context) error {
	return nil
}
