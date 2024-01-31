package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Queue(t *testing.T) {
	q := New()

	items := []*Item{
		{
			Type:  "type1",
			State: ItemStateNew,
		},
		{
			Type:  "type2",
			State: ItemStateFailed,
		},
	}

	q.Items = make([]*Item, 0)
	q.Items = append(q.Items, items...)

	assert.Len(t, q.GetItems(), 2)
	assert.Equal(t, 2, q.Total())
	assert.Equal(t, 1, q.Count(ItemStateNew))
	assert.Equal(t, 1, q.Count(ItemStateFailed))
	assert.Equal(t, 0, q.Count(ItemStateFiltered))

	assert.Equal(t, 1, q.CountByType("type1", ItemStateNew))
	assert.Equal(t, 0, q.CountByType("type1", ItemStatePending))

	assert.Equal(t, 1, q.CountByType("type2", ItemStateFailed))
	assert.Equal(t, 0, q.CountByType("type2", ItemStatePending))
}
