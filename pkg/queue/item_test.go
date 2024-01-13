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

func Test_Item(t *testing.T) {
	i := Item{
		Resource: &TestItemResource{},
		State:    ItemStateNew,
		Reason:   "brand new",
	}

	assert.Equal(t, ItemStateNew, i.GetState())
	assert.Equal(t, "brand new", i.GetReason())

	propVal, err := i.GetProperty("test")
	assert.NoError(t, err)
	assert.Equal(t, "testing", propVal)

	assert.True(t, i.Equals(i.Resource))
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
