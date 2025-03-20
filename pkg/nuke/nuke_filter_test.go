package nuke

import (
	"context"
	"flag"
	"io"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/scanner"
	"github.com/ekristen/libnuke/pkg/types"
)

func init() {
	if flag.Lookup("test.v") != nil {
		logrus.SetOutput(io.Discard)
	}
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)
}

func Test_NukeFiltersBad(t *testing.T) {
	filters := filter.Filters{
		TestResourceType: []filter.Filter{
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
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration2)

	filters := filter.Filters{
		TestResourceType2: []filter.Filter{
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

	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType2},
		Opts:          opts,
	})
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Scan(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_NukeFiltersMatchGroups_Match(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration2)

	filters := filter.Filters{
		TestResourceType2: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "test",
				Value:    "testing",
				Group:    "group1",
			},
			{
				Type:     filter.Exact,
				Property: "test2",
				Value:    "testing",
				Group:    "group2",
			},
		},
	}

	n := New(testParametersGroups, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne:     "testing",
		SecondResource: true,
	}
	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType2},
		Opts:          opts,
	})
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Scan(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_NukeFiltersMatchGroups_NoMatch(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration2)

	filters := filter.Filters{
		TestResourceType2: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "test",
				Value:    "testing",
				Group:    "group1",
			},
			{
				Type:     filter.Exact,
				Property: "test2",
				Value:    "testing!!!",
				Group:    "group2",
			},
		},
	}

	n := New(testParametersGroups, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne:     "testing",
		SecondResource: true,
	}
	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType2},
		Opts:          opts,
	})
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Scan(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_NukeFiltersMatchGroups_NoMatch_WithError(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration2)

	filters := filter.Filters{
		TestResourceType2: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "test",
				Value:    "testing",
				Group:    "group1",
			},
			{
				Type:     filter.Regex,
				Property: "test2",
				Value:    "^(testing$",
				Group:    "group2",
			},
		},
	}

	n := New(testParametersGroups, filters, nil)
	n.SetLogger(logrus.WithField("test", true))
	n.SetRunSleep(time.Millisecond * 5)

	opts := TestOpts{
		SessionOne:     "testing",
		SecondResource: true,
	}
	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType2},
		Opts:          opts,
	})
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Scan(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, "error parsing regexp: missing closing ): `^(testing$`", err.Error())
}

func Test_NukeFiltersMatchInverted(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration2)

	filters := filter.Filters{
		TestResourceType2: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "test",
				Value:    "testing",
				Invert:   true,
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
	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType2},
		Opts:          opts,
	})
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Scan(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_Nuke_Filters_NoMatch(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration2)

	filters := filter.Filters{
		TestResourceType: []filter.Filter{
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
	newScanner, err := scanner.New(&scanner.Config{
		Owner:         "Owner",
		ResourceTypes: []string{TestResourceType2},
		Opts:          opts,
	})
	assert.NoError(t, err)

	sErr := n.RegisterScanner(testScope, newScanner)
	assert.NoError(t, sErr)

	err = n.Scan(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFiltered))
}

func Test_Nuke_Filters_ErrorCustomProps(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(TestResourceRegistration)

	filters := filter.Filters{
		TestResourceType: []filter.Filter{
			{
				Type:     filter.Exact,
				Property: "Name",
				Value:    TestResourceType,
			},
		},
	}

	n := New(testParameters, filters, nil)
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

	err = n.Scan(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, 1, n.Queue.Total())
	assert.Equal(t, 1, n.Queue.Count(queue.ItemStateNew))
	assert.Equal(t, 0, n.Queue.Count(queue.ItemStateFiltered))
}

type TestResourceFilter struct {
	Props types.Properties
}

func (r *TestResourceFilter) Properties() types.Properties {
	return r.Props
}

func (r *TestResourceFilter) Remove(_ context.Context) error {
	return nil
}

func Test_Nuke_Filters_Extra(t *testing.T) {
	filters := filter.Filters{
		TestResourceType2: []filter.Filter{
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
		Resource: &TestResourceFilter{
			Props: types.Properties{
				"tag:aws:cloudformation:stack-name": "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081",
			},
		},
		Type: TestResourceType2,
	}

	err := n.Filter(i)
	assert.NoError(t, err)
	assert.Equal(t, i.Reason, "filtered by config")
}

func Test_Nuke_Filters_Filtered(t *testing.T) {
	cases := []struct {
		name      string
		error     bool
		resources []resource.Resource
		filters   filter.Filters
	}{
		{
			name: "exact",
			resources: []resource.Resource{
				&TestResourceFilter{
					Props: types.Properties{
						"tag:aws:cloudformation:stack-name": "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081",
					},
				},
			},
			filters: filter.Filters{
				TestResourceType2: []filter.Filter{
					{
						Type:     filter.Exact,
						Property: "tag:aws:cloudformation:stack-name",
						Value:    "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081",
					},
				},
			},
		},
		{
			name: "global",
			resources: []resource.Resource{
				&TestResourceFilter{
					Props: types.Properties{
						"tag:aws:cloudformation:stack-name": "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081",
					},
				},
			},
			filters: filter.Filters{
				filter.Global: []filter.Filter{
					{
						Type:     filter.Exact,
						Property: "tag:aws:cloudformation:stack-name",
						Value:    "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081",
					},
				},
				TestResourceType2: []filter.Filter{
					{
						Type:     filter.Exact,
						Property: "tag:testing",
						Value:    "test",
					},
				},
			},
		},
		{
			name:  "invalid",
			error: true,
			resources: []resource.Resource{
				&TestResourceFilter{
					Props: types.Properties{
						"tag:aws:cloudformation:stack-name": "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081",
					},
				},
			},
			filters: filter.Filters{
				TestResourceType2: []filter.Filter{
					{
						Type:     "invalid-filter",
						Property: "tag:aws:cloudformation:stack-name",
						Value:    "StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-c0bdd9c9-c338-4831-9c47-62443622c081",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := New(testParameters, tc.filters, nil)
			n.SetLogger(logrus.WithField("test", true))
			n.SetRunSleep(time.Millisecond * 5)

			for _, r := range tc.resources {
				i := &queue.Item{
					Resource: r,
					Type:     TestResourceType2,
				}

				err := n.Filter(i)
				if tc.error == true {
					assert.Error(t, err)
					continue
				}

				assert.NoError(t, err)
				assert.Equal(t, i.Reason, "filtered by config")
			}
		})
	}
}
