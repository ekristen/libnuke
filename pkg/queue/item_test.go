package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
	"github.com/ekristen/libnuke/pkg/unique"
)

func init() {
	logrus.SetLevel(logrus.TraceLevel)
}

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

func Test_ItemState_Stringer(t *testing.T) {
	cases := []struct {
		state ItemState
		want  string
	}{
		{
			state: ItemStateNew,
			want:  "new",
		},
		{
			state: ItemStatePending,
			want:  "pending",
		},
		{
			state: ItemStateNewDependency,
			want:  "new-dependency",
		},
		{
			state: ItemStatePendingDependency,
			want:  "pending-dependency",
		},
		{
			state: ItemStateWaiting,
			want:  "waiting",
		},
		{
			state: ItemStateFailed,
			want:  "failed",
		},
		{
			state: ItemStateFiltered,
			want:  "filtered",
		},
		{
			state: ItemStateFinished,
			want:  "finished",
		},
		{
			state: ItemStateHold,
			want:  "hold",
		},
		{
			state: ItemState(999),
			want:  "unknown",
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.state.String())
	}
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

type TestItemResourceUniqueKey struct {
	ID    string
	State string
}

func (r *TestItemResourceUniqueKey) Remove(_ context.Context) error {
	return nil
}
func (r *TestItemResourceUniqueKey) Properties() types.Properties {
	return types.NewProperties().Set("ID", r.ID).Set("State", r.State)
}
func (r *TestItemResourceUniqueKey) UniqueKey() string {
	return r.ID
}

func Test_ItemEqualUniqueKey(t *testing.T) {
	r1 := &TestItemResourceUniqueKey{
		ID:    "i-01b489457a60298dd",
		State: "running",
	}

	r2 := &TestItemResourceUniqueKey{
		ID:    "i-01b489457a60298dd", // Same ID
		State: "stopping",            // Different state (should be ignored)
	}

	i := &Item{Resource: r1}
	assert.True(t, i.Equals(r2), "Resources with same UniqueKey should be equal")

	r3 := &TestItemResourceUniqueKey{
		ID:    "i-1234567890abcdef0", // Different ID
		State: "running",             // Same state (should be ignored)
	}

	assert.False(t, i.Equals(r3), "Resources with different UniqueKey should not be equal")
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

// ------------------------------------------------------------------------

type TestGlobalHook struct {
	t  *testing.T
	tf func(t *testing.T, e *logrus.Entry)
}

func (h *TestGlobalHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TestGlobalHook) Fire(e *logrus.Entry) error {
	if h.tf != nil {
		h.tf(h.t, e)
	}

	return nil
}

type TestItemResourceLogger struct{}

func (r *TestItemResourceLogger) String() string {
	return "test"
}

func (r *TestItemResourceLogger) Remove(_ context.Context) error {
	return nil
}

func Test_ItemLoggerDefault(t *testing.T) {
	i := &Item{
		Resource: &TestItemResourceLogger{},
		State:    ItemStateNew,
		Reason:   "brand new",
		Type:     "TestResource",
		Owner:    "us-east-1",
	}

	i.Print()
}

func Test_ItemLoggerCustom(t *testing.T) {
	logger := logrus.New()
	defer func() {
		logger.ReplaceHooks(make(logrus.LevelHooks))
	}()

	hookCalled := false
	logger.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			hookCalled = true
			assert.Equal(t, "us-east-1", e.Data["owner"])
			assert.Equal(t, "TestResource", e.Data["type"])
			assert.Equal(t, 0, e.Data["state_code"])
			assert.Equal(t, "would remove", e.Message)
		},
	})

	i := &Item{
		Resource: &TestItemResourceLogger{},
		State:    ItemStateNew,
		Reason:   "brand new",
		Type:     "TestResource",
		Owner:    "us-east-1",
		Logger:   logger,
	}

	i.Print()

	assert.True(t, hookCalled)
}

// BenchmarkResource for single, double, and triple uniqueKey fields

type BenchmarkResourceOne struct {
	ID string `libnuke:"uniqueKey"`
}

func (r *BenchmarkResourceOne) UniqueKey() string {
	return r.ID
}

func (r *BenchmarkResourceOne) Remove(_ context.Context) error { return nil }

type BenchmarkResourceTwo struct {
	ID  string `libnuke:"uniqueKey"`
	Env string `libnuke:"uniqueKey"`
}

func (r *BenchmarkResourceTwo) Remove(_ context.Context) error { return nil }

func (r *BenchmarkResourceTwo) UniqueKey() string {
	return unique.Generate(r.ID, r.Env)
}

type BenchmarkResourceThree struct {
	ID   string `libnuke:"uniqueKey"`
	Env  string `libnuke:"uniqueKey"`
	Zone string `libnuke:"uniqueKey"`
}

func (r *BenchmarkResourceThree) UniqueKey() string {
	return unique.Generate(r.ID, r.Env, r.Zone)
}

func (r *BenchmarkResourceThree) Remove(_ context.Context) error { return nil }

func BenchmarkItemEquals(b *testing.B) {
	// 1 field uniqueKey
	r1 := &BenchmarkResourceOne{ID: "id-1"}
	r2 := &BenchmarkResourceOne{ID: "id-1"}
	r3 := &BenchmarkResourceOne{ID: "id-2"}
	item1 := &Item{Resource: r1}

	b.Run("Equals_OneField_Same", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			item1.Equals(r2)
		}
	})
	b.Run("Equals_OneField_Different", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			item1.Equals(r3)
		}
	})

	// 2 fields uniqueKey
	r4 := &BenchmarkResourceTwo{ID: "id-1", Env: "prod"}
	r5 := &BenchmarkResourceTwo{ID: "id-1", Env: "prod"}
	r6 := &BenchmarkResourceTwo{ID: "id-1", Env: "dev"}
	item2 := &Item{Resource: r4}

	b.Run("Equals_TwoField_Same", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			item2.Equals(r5)
		}
	})
	b.Run("Equals_TwoField_Different", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			item2.Equals(r6)
		}
	})

	// 3 fields uniqueKey
	r7 := &BenchmarkResourceThree{ID: "id-1", Env: "prod", Zone: "us-east-1a"}
	r8 := &BenchmarkResourceThree{ID: "id-1", Env: "prod", Zone: "us-east-1a"}
	r9 := &BenchmarkResourceThree{ID: "id-1", Env: "prod", Zone: "us-east-1b"}
	item3 := &Item{Resource: r7}

	b.Run("Equals_ThreeField_Same", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			item3.Equals(r8)
		}
	})
	b.Run("Equals_ThreeField_Different", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			item3.Equals(r9)
		}
	})
}
