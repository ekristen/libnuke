package nuke

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/scan"
)

type TestResourceSuccess struct {
}

func (r *TestResourceSuccess) Remove(_ context.Context) error { return nil }
func (r *TestResourceSuccess) String() string                 { return "TestResourceSuccess" }

type TestResourceSuccessLister struct {
	listed bool
}

func (l *TestResourceSuccessLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	if l.listed {
		return []resource.Resource{}, nil
	}
	l.listed = true
	return []resource.Resource{&TestResourceSuccess{}}, nil
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

	resource.ClearRegistry()
	resource.Register(&resource.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	scannerErr := n.RegisterScanner(testScope, scan.NewScanner("Owner", []string{"TestResourceSuccess"}, nil))
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateNew))
	assert.Equal(t, 1, n.Queue.Total())
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

	resource.ClearRegistry()
	resource.Register(&resource.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	scannerErr := n.RegisterScanner(testScope, scan.NewScanner("Owner", []string{"TestResourceSuccess"}, nil))
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

	scannerErr := n.RegisterScanner(testScope, scan.NewScanner("Owner", []string{"TestResource4"}, nil))
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFinished))
}

// Test_Nuke_Run_Failure tests a run with a resource that fails to remove, so it should be in the failed state.
// It also tests that a resource is successfully removed as well, to test the entire fail state.
func Test_Nuke_Run_Failure(t *testing.T) {
	n := New(testParametersRemove, nil, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	resource.ClearRegistry()
	resource.Register(&resource.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	resource.Register(&resource.Registration{
		Name:   "TestResourceFailure",
		Lister: &TestResourceFailureLister{},
	})

	newScanner := scan.NewScanner("Owner", []string{"TestResourceSuccess", "TestResourceFailure"}, nil)
	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.Error(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFinished))
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFailed))
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

	resource.ClearRegistry()
	resource.Register(&resource.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceWaitLister{},
	})

	newScanner := scan.NewScanner("Owner", []string{"TestResourceSuccess"}, nil)
	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.Error(t, runErr)
	assert.Equal(t, "max wait retries of 3 exceeded", runErr.Error())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateWaiting))
}

// ---------------------

type TestResourceAlpha struct {
}

func (r *TestResourceAlpha) Remove(_ context.Context) error { return nil }
func (r *TestResourceAlpha) String() string                 { return "TestResourceAlpha" }

type TestResourceAlphaLister struct {
	listed bool
}

func (l *TestResourceAlphaLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	if l.listed {
		return []resource.Resource{}, nil
	}
	l.listed = true
	return []resource.Resource{&TestResourceAlpha{}}, nil
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

	resource.ClearRegistry()
	resource.Register(&resource.Registration{
		Name:   "TestResourceAlpha",
		Lister: &TestResourceAlphaLister{},
	})
	resource.Register(&resource.Registration{
		Name:   "TestResourceBeta",
		Lister: &TestResourceAlphaLister{},
		DependsOn: []string{
			"TestResourceAlpha",
		},
	})

	newScanner := scan.NewScanner("Owner", []string{"TestResourceAlpha", "TestResourceBeta"}, nil)
	scannerErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run(context.TODO())
	assert.NoError(t, runErr)

	assert.Equal(t, 2, n.Queue.Count(queue.ItemStateFinished))
}
