package nuke

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

func Test_NukeFiltersBad(t *testing.T) {
	filters := filter.Filters{
		testResourceType: []filter.Filter{
			{
				Type: filter.Exact,
			},
		},
	}

	n := New(testParameters, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	err := n.Run(context.TODO())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "testResourceType: has an invalid filter")
}

func Test_NukeFiltersMatch(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration2)

	filters := filter.Filters{
		testResourceType2: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "test",
				Value:    "testing",
			},
		},
	}

	n := New(testParameters, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne:     "testing",
		SecondResource: true,
	}
	scanner := NewScanner("owner", []string{testResourceType2}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Scan(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_NukeFiltersMatchInverted(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration2)

	filters := filter.Filters{
		testResourceType2: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "test",
				Value:    "testing",
				Invert:   "true",
			},
		},
	}

	n := New(testParameters, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne:     "testing",
		SecondResource: true,
	}
	scanner := NewScanner("owner", []string{testResourceType2}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Scan(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_Nuke_Filters_NoMatch(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration2)

	filters := filter.Filters{
		testResourceType: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "test",
				Value:    "testing",
			},
		},
	}

	n := New(testParameters, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne:     "testing",
		SecondResource: true,
	}
	scanner := NewScanner("owner", []string{testResourceType2}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Scan(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_Nuke_Filters_ErrorCustomProps(t *testing.T) {
	resource.ClearRegistry()
	resource.Register(testResourceRegistration)

	filters := filter.Filters{
		testResourceType: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "Name",
				Value:    testResourceType,
			},
		},
	}

	n := New(testParameters, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne: "testing",
	}
	scanner := NewScanner("owner", []string{testResourceType}, opts)

	sErr := n.RegisterScanner(testScope, scanner)
	assert.NoError(t, sErr)

	err := n.Scan(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, "*nuke.TestResource does not support custom properties", err.Error())
}

type TestResourceFilter struct {
}

func (r *TestResourceFilter) Properties() types.Properties {
	props := types.NewProperties()

	tagName := ptr.String("aws:cloudformation:stack-name")
	tagVal := "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081"

	props.SetTag(tagName, tagVal)
	return props
}

func (r *TestResourceFilter) Remove(_ context.Context) error {
	return nil
}

func Test_Nuke_Filters_Extra(t *testing.T) {
	filters := filter.Filters{
		testResourceType2: []filter.Filter{
			{
				Type:     filter.Glob,
				Property: "tag:aws:cloudformation:stack-name",
				Value:    "StackSet-AWSControlTowerBP*",
			},
		},
	}

	n := New(testParameters, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	i := &queue.Item{
		Resource: &TestResourceFilter{},
		Type:     testResourceType2,
	}

	err := n.Filter(i)
	assert.NoError(t, err)
	assert.Equal(t, i.Reason, "filtered by config")
}
