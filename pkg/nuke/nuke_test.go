package nuke

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/resource"
)

var testParameters = Parameters{
	Force:      true,
	ForceSleep: 3,
	Quiet:      true,
}

var testParametersRemove = Parameters{
	Force:      true,
	ForceSleep: 3,
	Quiet:      true,
	NoDryRun:   true,
}

const testScope resource.Scope = "test"

func Test_Nuke_Version(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	n.RegisterVersion("1.0.0-test")

	// Redirect stdout to a buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the Version function
	n.Version()

	// Restore stdout
	os.Stdout = old

	w.Close()

	out, _ := io.ReadAll(r)
	outString := string(out)

	// Check the output
	if !strings.Contains(outString, "1.0.0-test") {
		t.Errorf("Version() = %v, want %v", out, "1.0.0-test")
	}
}

func Test_Nuke_FeatureFlag(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	n.RegisterFeatureFlags("test", ptr.Bool(true), ptr.Bool(true))

	flag, err := n.FeatureFlags.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, true, flag.Enabled())

	flag1, err := n.FeatureFlags.Get("testing")
	assert.Error(t, err)
	assert.Nil(t, flag1)
}

func Test_Nuke_Validators_Default(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	err := n.Validate()
	assert.NoError(t, err)
}

func Test_Nuke_Validators_Register1(t *testing.T) {
	n := New(testParameters, nil)
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
	n := New(testParameters, nil)
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
	p := Parameters{
		Force:      true,
		ForceSleep: 1,
		Quiet:      true,
	}
	n := New(p, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	err := n.Validate()
	assert.Error(t, err)
	assert.Equal(t, "value for --force-sleep cannot be less than 3 seconds. This is for your own protection", err.Error())
}

func Test_Nuke_ResourceTypes(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	n.RegisterResourceTypes(testScope, "TestResource")

	assert.Len(t, n.ResourceTypes[testScope], 1)
}

func Test_Nuke_Scanners(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := struct {
		name string
	}{
		name: "test",
	}

	s := NewScanner("test", []string{"TestResource"}, opts)

	err := n.RegisterScanner(testScope, s)
	assert.NoError(t, err)

	assert.Len(t, n.Scanners[testScope], 1)
}

func Test_Nuke_Scanners_Duplicate(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := struct {
		name string
	}{
		name: "test",
	}

	s := NewScanner("test", []string{"TestResource"}, opts)

	err := n.RegisterScanner(testScope, s)
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, s)
	assert.Error(t, sErr)

	assert.Len(t, n.Scanners[testScope], 1)
}

func Test_Nuke_RegisterPrompt(t *testing.T) {
	n := New(testParameters, nil)
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
	resource.ClearRegistry()
	resource.Register(testResourceRegistration)
	resource.Register(resource.Registration{
		Name:  testResourceType2,
		Scope: "account",
		Lister: TestResourceLister{
			Filtered: true,
		},
	})

	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne: "testing",
	}
	scanner := NewScanner("owner", []string{testResourceType, testResourceType2}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Scan()
	assert.NoError(t, err)

	assert.Equal(t, 2, n.Queue.Total())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateNew))
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFiltered))
}

// ---------------------------------------------------------------------

type TestResource3 struct {
	Error bool
}

func (r TestResource3) Remove() error {
	if r.Error {
		return fmt.Errorf("remove error")
	}
	return nil
}

func Test_Nuke_HandleRemove(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	i := &queue.Item{
		Resource: &TestResource3{},
		State:    queue.ItemStateNew,
	}

	n.HandleRemove(i)
	assert.Equal(t, queue.ItemStatePending, i.State)
}

func Test_Nuke_HandleRemoveError(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	i := &queue.Item{
		Resource: &TestResource3{
			Error: true,
		},
		State: queue.ItemStateNew,
	}

	n.HandleRemove(i)
	assert.Equal(t, queue.ItemStateFailed, i.State)
}

// ------------------------------------------------------------

func Test_Nuke_Run(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration)

	p := Parameters{
		Force:      true,
		ForceSleep: 3,
		Quiet:      true,
		NoDryRun:   true,
	}

	n := New(p, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne: "testing",
	}
	scanner := NewScanner("owner", []string{testResourceType}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Run()
	assert.NoError(t, err)
}

func Test_Nuke_Run_Error(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(resource.Registration{
		Name:  testResourceType2,
		Scope: "account",
		Lister: TestResourceLister{
			RemoveError: true,
		},
	})

	p := Parameters{
		Force:      true,
		ForceSleep: 3,
		Quiet:      true,
		NoDryRun:   true,
	}
	n := New(p, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne: "testing",
	}
	scanner := NewScanner("owner", []string{testResourceType2}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Run()
	assert.NoError(t, err)
}

// ------------------------------------------------------------

var TestResource4Resources []resource.Resource

type TestResource4 struct {
	id       string
	parentID string
}

func (r *TestResource4) Remove() error {
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

func (l *TestResource4Lister) List(o interface{}) ([]resource.Resource, error) {
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
	n := New(testParametersRemove, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	resource.ClearRegistry()
	resource.Register(resource.Registration{
		Name:   "TestResource4",
		Scope:  testScope,
		Lister: &TestResource4Lister{},
	})

	scannerErr := n.RegisterScanner(testScope, NewScanner("owner", []string{"TestResource4"}, nil))
	assert.NoError(t, scannerErr)

	runErr := n.Run()
	assert.NoError(t, runErr)

	assert.Equal(t, 5, n.Queue.Count(queue.ItemStateFinished))
}

// -----------------------------------------------
