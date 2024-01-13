package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/types"
)

type TestItemResource struct{}

func (r TestItemResource) Properties() types.Properties {
	props := types.NewProperties()
	props.Set("test", "testing")
	return props
}

func (r TestItemResource) Remove() error {
	return nil
}

func (r TestItemResource) String() string {
	return "test"
}

type TestItemResource2 struct{}

func (r TestItemResource2) Remove() error {
	return nil
}

var testItem = Item{
	Resource: &TestItemResource{},
	State:    ItemStateNew,
	Reason:   "brand new",
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

	propVal, err := i.GetProperty("test")
	assert.NoError(t, err)
	assert.Equal(t, "testing", propVal)

	assert.True(t, i.Equals(i.Resource))

	assert.False(t, i.Equals(testItem2.Resource))
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			i := &Item{
				Resource: &TestItemResource{},
				State:    tc.state,
				Type:     "TestResource",
				Owner:    "us-east-1",
			}
			i.Print()
		})
	}
}
