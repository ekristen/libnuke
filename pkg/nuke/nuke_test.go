package nuke

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/featureflag"
	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

var testParameters = Parameters{
	Force:      true,
	ForceSleep: 3,
	Quiet:      true,
}

const testScope resource.Scope = "test"

func Test_Nuke_Version(t *testing.T) {
	n := New(testParameters, nil)
	n.SetLogger(logrus.WithField("test", true))

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
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

	n.RegisterFeatureFlags("test", ptr.Bool(true), ptr.Bool(true))

	flag, err := n.FeatureFlags.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, true, flag.Enabled())

	flag1, err := n.FeatureFlags.Get("testing")
	assert.Error(t, err)
	assert.Nil(t, flag1)
}

func Test_Nuke_Validators_Default(t *testing.T) {
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

	err := n.Validate()
	assert.NoError(t, err)
}

func Test_Nuke_Validators_Register1(t *testing.T) {
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

	n.RegisterValidateHandler(func() error {
		return fmt.Errorf("validator called")
	})

	err := n.Validate()
	assert.Error(t, err)
	assert.Equal(t, "validator called", err.Error())
}

func Test_Nuke_Validators_Register2(t *testing.T) {
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

	n.RegisterValidateHandler(func() error {
		return fmt.Errorf("validator called")
	})

	n.RegisterValidateHandler(func() error {
		return fmt.Errorf("second validator called")
	})

	assert.Len(t, n.ValidateHandlers, 2)
}

func Test_Nuke_Validators_Error(t *testing.T) {
	n := &Nuke{
		Parameters: Parameters{
			Force:      true,
			ForceSleep: 1,
			Quiet:      true,
		},
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

	err := n.Validate()
	assert.Error(t, err)
	assert.Equal(t, "value for --force-sleep cannot be less than 3 seconds. This is for your own protection", err.Error())
}

func Test_Nuke_ResourceTypes(t *testing.T) {
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

	n.RegisterResourceTypes(testScope, "TestResource")

	assert.Len(t, n.ResourceTypes[testScope], 1)
}

func Test_Nuke_Scanners(t *testing.T) {
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

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
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

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
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

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

	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

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

func Test_Nuke_Filters_Match(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration2)

	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		Filters: filter.Filters{
			testResourceType2: []filter.Filter{
				{
					Type:     filter.Exact,
					Property: "test",
					Value:    "testing",
				},
			},
		},
		log: logrus.WithField("test", true),
	}

	opts := TestOpts{
		SessionOne:     "testing",
		SecondResource: true,
	}
	scanner := NewScanner("owner", []string{testResourceType2}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Scan()
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_Nuke_Filters_NoMatch(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration2)

	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		Filters: filter.Filters{
			testResourceType: []filter.Filter{
				{
					Type:     filter.Exact,
					Property: "test",
					Value:    "testing",
				},
			},
		},
		log: logrus.WithField("test", true),
	}

	opts := TestOpts{
		SessionOne:     "testing",
		SecondResource: true,
	}
	scanner := NewScanner("owner", []string{testResourceType2}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Scan()
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_Nuke_Filters_ErrorCustomProps(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration)

	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		Filters: filter.Filters{
			testResourceType: []filter.Filter{
				{
					Type:     filter.Exact,
					Property: "Name",
					Value:    testResourceType,
				},
			},
		},
		log: logrus.WithField("test", true),
	}

	opts := TestOpts{
		SessionOne: "testing",
	}
	scanner := NewScanner("owner", []string{testResourceType}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Scan()
	assert.Error(t, err)
	assert.Equal(t, "*nuke.TestResource does not support custom properties", err.Error())
}

type TestResourceFilter struct {
}

func (r TestResourceFilter) Properties() types.Properties {
	props := types.NewProperties()

	tagName := ptr.String("aws:cloudformation:stack-name")
	tagVal := "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081"

	props.SetTag(tagName, tagVal)
	return props
}

func (r TestResourceFilter) Remove() error {
	return nil
}

func Test_Nuke_Filters_Extra(t *testing.T) {
	n := &Nuke{
		Parameters:   testParameters,
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		Filters: filter.Filters{
			testResourceType2: []filter.Filter{
				{
					Type:     filter.Glob,
					Property: "tag:aws:cloudformation:stack-name",
					Value:    "StackSet-AWSControlTowerBP*",
				},
			},
		},
		log: logrus.WithField("test", true),
	}

	i := &queue.Item{
		Resource: &TestResourceFilter{},
		Type:     testResourceType2,
	}

	err := n.Filter(i)
	assert.NoError(t, err)
	assert.Equal(t, i.Reason, "filtered by config")
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
	n := &Nuke{
		Parameters: testParameters,
		Queue:      queue.Queue{},
		log:        logrus.WithField("test", true),
	}

	i := &queue.Item{
		Resource: &TestResource3{},
		State:    queue.ItemStateNew,
	}

	n.HandleRemove(i)
	assert.Equal(t, queue.ItemStatePending, i.State)
}

func Test_Nuke_HandleRemoveError(t *testing.T) {
	n := &Nuke{
		Parameters: testParameters,
		Queue:      queue.Queue{},
		log:        logrus.WithField("test", true),
	}

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

	n := &Nuke{
		Parameters: Parameters{
			Force:      true,
			ForceSleep: 3,
			Quiet:      true,
			NoDryRun:   true,
		},
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

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

	n := &Nuke{
		Parameters: Parameters{
			Force:      true,
			ForceSleep: 3,
			Quiet:      true,
			NoDryRun:   true,
		},
		Queue:        queue.Queue{},
		FeatureFlags: &featureflag.FeatureFlags{},
		log:          logrus.WithField("test", true),
	}

	opts := TestOpts{
		SessionOne: "testing",
	}
	scanner := NewScanner("owner", []string{testResourceType2}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Run()
	assert.NoError(t, err)
}
