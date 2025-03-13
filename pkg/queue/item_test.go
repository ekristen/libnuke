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

type TestKeyedResource struct {
	ID    *string `libnuke:"nonRepeatableKey"`
	State *string
}

func (r *TestKeyedResource) Remove(_ context.Context) error {
	return nil
}

func Test_ItemEqualNonRepeatableKeys(t *testing.T) {
	id1 := "i-01b489457a60298dd"
	state1 := "running"
	r1 := &TestKeyedResource{ID: &id1, State: &state1}

	id2 := "i-01b489457a60298dd" // Same ID
	state2 := "stopping"         // Different state (should be ignored)
	r2 := &TestKeyedResource{ID: &id2, State: &state2}

	i := &Item{Resource: r1}
	assert.True(t, i.Equals(r2), "Resources with same nonRepeatableKey should be equal")

	// Test with non-matching nonRepeatableKeys
	id3 := "i-1234567890abcdef0" // Different ID
	state3 := "running"          // Same state (should be ignored)
	r3 := &TestKeyedResource{ID: &id3, State: &state3}

	assert.False(t, i.Equals(r3), "Resources with different nonRepeatableKey should not be equal")
}

// ------------------------------------------------------------------------

type TestKeyedResourceWithTags struct {
	ID    *string `property:"name=id" libnuke:"nonRepeatableKey,futureValue"`
	State *string `property:"name=state"`
}

func (r *TestKeyedResourceWithTags) Remove(_ context.Context) error {
	return nil
}

func Test_ItemEqualNonRepeatableKeyWithTags(t *testing.T) {
	id1 := "i-01b489457a60298dd"
	state1 := "running"
	r1 := &TestKeyedResourceWithTags{ID: &id1, State: &state1}

	id2 := "i-01b489457a60298dd" // Same ID
	state2 := "stopping"         // Different state (should be ignored)
	r2 := &TestKeyedResourceWithTags{ID: &id2, State: &state2}

	i := &Item{Resource: r1}
	assert.True(t, i.Equals(r2), "Should find nonRepeatableKey even when mixed with other tags")
}

// ------------------------------------------------------------------------

type TestMultiKeyedResource struct {
	Name         *string    `libnuke:"nonRepeatableKey"`
	CreationTime *time.Time `libnuke:"nonRepeatableKey"`
	LastEvent    *time.Time
}

func (r *TestMultiKeyedResource) Remove(_ context.Context) error {
	return nil
}

func Test_ItemEqualMultipleNonRepeatableKeys(t *testing.T) {
	now := time.Now()

	name1 := "TestLogGroup"
	creationTime1 := now
	lastEvent1 := now
	r1 := &TestMultiKeyedResource{Name: &name1, CreationTime: &creationTime1, LastEvent: &lastEvent1}

	// All keys match
	name2 := "TestLogGroup"
	creationTime2 := now
	lastEvent2 := now.Add(1 * time.Hour) // Should be ignored
	r2 := &TestMultiKeyedResource{Name: &name2, CreationTime: &creationTime2, LastEvent: &lastEvent2}

	i := &Item{Resource: r1}
	assert.True(t, i.Equals(r2), "Resources with all matching keys should be equal")

	// One key doesn't match
	name3 := "TestLogGroup"
	creationTime3 := now.Add(1 * time.Hour) // Different creation time (e.g. resource nuked but recreated before run finishes)
	lastEvent3 := now.Add(1 * time.Hour)    // Should be ignored
	r3 := &TestMultiKeyedResource{Name: &name3, CreationTime: &creationTime3, LastEvent: &lastEvent3}

	assert.False(t, i.Equals(r3), "Resources with any non-matching key should not be equal")
}

// ------------------------------------------------------------------------

type TestComplexKeyResource struct {
	Name *string            `libnuke:"nonRepeatableKey"`
	Tags map[string]*string `libnuke:"nonRepeatableKey"`
}

func (r *TestComplexKeyResource) Remove(_ context.Context) error {
	return nil
}

func Test_ItemEqualComplexNonRepeatableKeys(t *testing.T) {
	name1 := "resource-123"
	tag1Value := "value1"
	tag2Value := "value2"
	tag3Value := "value3"

	r1 := &TestComplexKeyResource{
		Name: &name1,
		Tags: map[string]*string{
			"tag1": &tag1Value,
			"tag2": &tag2Value,
		},
	}

	// Same complex key values
	name2 := "resource-123"
	r2 := &TestComplexKeyResource{
		Name: &name2,
		Tags: map[string]*string{
			"tag1": &tag1Value,
			"tag2": &tag2Value,
		},
	}

	// Different map contents
	name3 := "resource-123"
	r3 := &TestComplexKeyResource{
		Name: &name3,
		Tags: map[string]*string{
			"tag1": &tag1Value,
			"tag3": &tag3Value,
		},
	}

	i := &Item{Resource: r1}
	assert.True(t, i.Equals(r2), "Resources with same complex nonRepeatableKey should be equal")
	assert.False(t, i.Equals(r3), "Resources with different complex nonRepeatableKey should not be equal")
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
