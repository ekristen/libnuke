package nuke

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/resource"
)

type TestResourceSuccess struct{}

func (r *TestResourceSuccess) Remove() error  { return nil }
func (r *TestResourceSuccess) String() string { return "TestResourceFailure" }

type TestResourceSuccessLister struct {
	listed bool
}

func (l *TestResourceSuccessLister) List(o interface{}) ([]resource.Resource, error) {
	if l.listed {
		return []resource.Resource{}, nil
	}
	l.listed = true
	return []resource.Resource{&TestResourceSuccess{}}, nil
}

type TestResourceFailure struct{}

func (r *TestResourceFailure) Remove() error  { return fmt.Errorf("unable to remove") }
func (r *TestResourceFailure) String() string { return "TestResourceFailure" }

type TestResourceFailureLister struct{}

func (l *TestResourceFailureLister) List(o interface{}) ([]resource.Resource, error) {
	return []resource.Resource{&TestResourceFailure{}}, nil
}

type TestResourceWaitLister struct{}

func (l *TestResourceWaitLister) List(o interface{}) ([]resource.Resource, error) {
	return []resource.Resource{&TestResourceSuccess{}}, nil
}

// Test_Nuke_Run_SimpleWithNoDryRun tests a simple run with no dry run enabled so all resources are removed.
func Test_Nuke_Run_SimpleWithNoDryRun(t *testing.T) {
	n := New(testParametersRemove, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	scannerErr := n.RegisterScanner(testScope, NewScanner("owner", []string{"TestResource4"}, nil))
	assert.NoError(t, scannerErr)

	runErr := n.Run()
	assert.NoError(t, runErr)

	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFinished))
}

// Test_Nuke_Run_Failure tests a run with a resource that fails to remove, so it should be in the failed state.
// It also tests that a resource is successfully removed as well, to test the entire fail state.
func Test_Nuke_Run_Failure(t *testing.T) {
	n := New(testParametersRemove, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	resource.ClearRegistry()
	resource.Register(resource.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceSuccessLister{},
	})

	resource.Register(resource.Registration{
		Name:   "TestResourceFailure",
		Lister: &TestResourceFailureLister{},
	})

	scanner := NewScanner("owner", []string{"TestResourceSuccess", "TestResourceFailure"}, nil)
	scannerErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run()
	assert.Error(t, runErr)

	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFinished))
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFailed))
}

var testParametersMaxWaitRetries = Parameters{
	Force:          true,
	ForceSleep:     3,
	Quiet:          true,
	NoDryRun:       true,
	MaxWaitRetries: 3,
}

func Test_NukeRunWithMaxWaitRetries(t *testing.T) {
	n := New(testParametersMaxWaitRetries, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	resource.ClearRegistry()
	resource.Register(resource.Registration{
		Name:   "TestResourceSuccess",
		Lister: &TestResourceWaitLister{},
	})

	scanner := NewScanner("owner", []string{"TestResourceSuccess"}, nil)
	scannerErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, scannerErr)

	runErr := n.Run()
	assert.Error(t, runErr)
	assert.Equal(t, "max wait retries of 3 exceeded", runErr.Error())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateWaiting))
}
