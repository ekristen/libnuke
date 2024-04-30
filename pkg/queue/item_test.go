package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

type TestItemResource struct {
	id string
}

func (r *TestItemResource) Properties() types.Properties {
	props := types.NewProperties()
	props.Set("name", r.id)
	return props
}
func (r *TestItemResource) Remove(_ context.Context) error {
	return nil
}
func (r *TestItemResource) String() string {
	return r.id
}

type TestItemResource2 struct{}

func (r TestItemResource2) Remove(_ context.Context) error {
	return nil
}

type TestItemResourceLister struct{}

func (l *TestItemResourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	return []resource.Resource{&TestItemResource{id: "test"}}, nil
}

var testItem = Item{
	Resource: &TestItemResource{id: "test"},
	State:    ItemStateNew,
	Reason:   "brand new",
	Type:     "TestResource",
}

var testItem2 = Item{
	Resource: &TestItemResource2{},
	State:    ItemStateNew,
	Reason:   "brand new",
}

func Test_Item(t *testing.T) {
	i := testItem

	assert.Equal(t, ItemStateNew, i.GetState())
	assert.Equal(t, "brand new", i.GetReason())

	propVal, err := i.GetProperty("name")
	assert.NoError(t, err)
	assert.Equal(t, "test", propVal)

	assert.True(t, i.Equals(i.Resource))
	assert.False(t, i.Equals(testItem2.Resource))
}

func Test_ItemList(t *testing.T) {
	registry.ClearRegistry()
	registry.Register(&registry.Registration{
		Name:   "TestResource",
		Lister: &TestItemResourceLister{},
	})

	i := testItem
	list, err := i.List(context.TODO(), nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
}

func Test_Item_LegacyStringer(t *testing.T) {
	i := testItem
	val, err := i.GetProperty("")
	assert.NoError(t, err)
	assert.Equal(t, "test", val)
}

func Test_Item_LegacyStringer_NoSupport(t *testing.T) {
	i := testItem2
	_, err := i.GetProperty("")
	assert.Error(t, err)
	assert.Equal(t, "*queue.TestItemResource2 does not support legacy IDs", err.Error())
}

func Test_Item_Properties_NoSupport(t *testing.T) {
	i := testItem2
	_, err := i.GetProperty("test-prop")
	assert.Error(t, err)
	assert.Equal(t, "*queue.TestItemResource2 does not support custom properties", err.Error())
}

func Test_ItemPrint(t *testing.T) {
	cases := []struct {
		name  string
		state ItemState
		want  string
	}{
		{
			name:  "new",
			state: ItemStateNew,
			want:  "would remove",
		},
		{
			name:  "pending",
			state: ItemStatePending,
			want:  "triggered remove",
		},
		{
			name:  "new-dependency",
			state: ItemStateNewDependency,
			want:  "would remove after dependencies",
		},
		{
			name:  "pending-dependency",
			state: ItemStatePendingDependency,
			want:  "waiting on dependencies (brand new)",
		},
		{
			name:  "waiting",
			state: ItemStateWaiting,
			want:  "waiting",
		},
		{
			name:  "failed",
			state: ItemStateFailed,
			want:  "failed",
		},
		{
			name:  "filtered",
			state: ItemStateFiltered,
			want:  "varies",
		},
		{
			name:  "finished",
			state: ItemStateFinished,
			want:  "finished",
		},
		{
			name:  "hold",
			state: ItemStateHold,
			want:  "waiting for parent removal",
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			i := &Item{
				Resource: &TestItemResource{
					id: fmt.Sprintf("test%d", i),
				},
				State: tc.state,
				Type:  "TestResource",
				Owner: "us-east-1",
			}
			i.Print()
		})
	}
}

// ------------------------------------------------------------------------

type TestItemResourceProperties struct{}

func (r *TestItemResourceProperties) Remove(_ context.Context) error {
	return nil
}
func (r *TestItemResourceProperties) Properties() types.Properties {
	return types.NewProperties().Set("test", "testing")
}

func Test_ItemEqualProperties(t *testing.T) {
	i := &Item{
		Resource: &TestItemResourceProperties{},
		State:    ItemStateNew,
		Reason:   "brand new",
		Type:     "TestResource",
	}

	assert.True(t, i.Equals(i.Resource))
}

// ------------------------------------------------------------------------

type TestItemResourceStringer struct{}

func (r *TestItemResourceStringer) Remove(_ context.Context) error {
	return nil
}
func (r *TestItemResourceStringer) String() string {
	return "test"
}

func Test_ItemEqualStringer(t *testing.T) {
	i := &Item{
		Resource: &TestItemResourceStringer{},
		State:    ItemStateNew,
		Reason:   "brand new",
		Type:     "TestResource",
	}

	ni := &Item{
		Resource: &TestItemResourceNothing{},
		State:    ItemStateNew,
		Reason:   "brand new",
		Type:     "TestResource",
	}

	assert.True(t, i.Equals(i.Resource))
	assert.False(t, i.Equals(ni.Resource))
}

// ------------------------------------------------------------------------

type TestItemResourceNothing struct{}

func (r *TestItemResourceNothing) Remove(_ context.Context) error {
	return nil
}

func Test_ItemEqualNothing(t *testing.T) {
	i := &Item{
		Resource: &TestItemResourceNothing{},
		State:    ItemStateNew,
		Reason:   "brand new",
		Type:     "TestResource",
	}

	assert.False(t, i.Equals(i.Resource))
}

// ------------------------------------------------------------------------

type TestItemResourceRevenant struct {
	Props types.Properties
}

func (r *TestItemResourceRevenant) Remove(_ context.Context) error {
	return nil
}
func (r *TestItemResourceRevenant) Properties() types.Properties {
	return r.Props
}

func Test_ItemRevenant(t *testing.T) {
	i := &Item{
		Resource: &TestItemResourceRevenant{
			Props: types.NewProperties().Set("CreatedAt", time.Now().UTC()),
		},
		State:  ItemStateNew,
		Reason: "brand new",
		Type:   "TestResource",
	}

	j := &Item{
		Resource: &TestItemResourceRevenant{
			Props: types.NewProperties().Set("CreatedAt", time.Now().UTC().Add(4*time.Minute)),
		},
		State:  ItemStateNew,
		Reason: "brand new",
		Type:   "TestResource",
	}

	assert.False(t, j.Equals(i.Resource))
}
