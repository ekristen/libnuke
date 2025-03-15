package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/types"
)

func TestPropertiesEquals(t *testing.T) {
	cases := []struct {
		p1, p2 types.Properties
		result bool
	}{
		{
			p1:     nil,
			p2:     nil,
			result: true,
		},
		{
			p1:     nil,
			p2:     types.NewProperties(),
			result: false,
		},
		{
			p1:     types.NewProperties(),
			p2:     types.NewProperties(),
			result: true,
		},
		{
			p1:     types.NewProperties().Set("blub", "blubber"),
			p2:     types.NewProperties().Set("blub", "blubber"),
			result: true,
		},
		{
			p1:     types.NewProperties().Set("blub", "foo"),
			p2:     types.NewProperties().Set("blub", "bar"),
			result: false,
		},
		{
			p1:     types.NewProperties().Set("bim", "baz").Set("blub", "blubber"),
			p2:     types.NewProperties().Set("bim", "baz").Set("blub", "blubber"),
			result: true,
		},
		{
			p1:     types.NewProperties().Set("bim", "baz").Set("blub", "foo"),
			p2:     types.NewProperties().Set("bim", "baz").Set("blub", "bar"),
			result: false,
		},
		{
			p1:     types.NewProperties().Set("bim", "baz").Set("blub", "foo"),
			p2:     types.NewProperties().Set("bim", "baz1"),
			result: false,
		},
		{
			p1:     types.NewProperties().Set("blub", "foo"),
			p2:     types.NewProperties().Set("bim", "baz1"),
			result: false,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if tc.p1.Equals(tc.p2) != tc.result {
				t.Errorf("Test Case failed. Want %t. Got %t.", !tc.result, tc.result)
				t.Errorf("p1: %s", tc.p1.String())
				t.Errorf("p2: %s", tc.p2.String())
			} else if tc.p2.Equals(tc.p1) != tc.result {
				t.Errorf("Test Case reverse check failed. Want %t. Got %t.", !tc.result, tc.result)
				t.Errorf("p1: %s", tc.p1.String())
				t.Errorf("p2: %s", tc.p2.String())
			}
		})
	}
}

func TestPropertiesSetGet(t *testing.T) {
	var nilString *string = nil
	var nilBool *bool = nil
	var nilInt *int = nil
	var nilInt64 *int64 = nil

	cases := []struct {
		name  string
		key   *string
		value interface{}
		want  string
	}{
		{
			name:  "string",
			key:   ptr.String("name"),
			value: "blubber",
			want:  "blubber",
		},
		{
			name:  "string_ptr",
			key:   ptr.String("name_string"),
			value: ptr.String("blubber"),
			want:  "blubber",
		},
		{
			name:  "int",
			key:   ptr.String("integer"),
			value: 42,
			want:  "42",
		},
		{
			name:  "int_ptr",
			key:   ptr.String("int_ptr"),
			value: &[]int{32}[0],
			want:  "32",
		},
		{
			name:  "int64",
			key:   ptr.String("int64"),
			value: []int64{64}[0],
			want:  "64",
		},
		{
			name:  "int64_ptr",
			key:   ptr.String("int64_ptr"),
			value: &[]int64{64}[0],
			want:  "64",
		},
		{
			name:  "byte",
			key:   ptr.String("byte"),
			value: []byte("testbyte"),
			want:  "testbyte",
		},
		{
			name:  "bool",
			key:   ptr.String("bool"),
			value: true,
			want:  "true",
		},
		{
			name:  "bool_ptr",
			key:   ptr.String("bool_ptr"),
			value: &[]bool{false}[0],
			want:  "false",
		},
		{
			name:  "nil",
			key:   ptr.String("nothing"),
			value: nil,
			want:  "",
		},
		{
			name:  "empty_key",
			key:   ptr.String(""),
			value: "empty",
			want:  "empty",
		},
		{
			name:  "nil_key",
			key:   nil,
			value: "empty",
			want:  "empty",
		},
		{
			name:  "nil_string",
			key:   ptr.String("nil_string"),
			value: nilString,
			want:  "",
		},
		{
			name:  "nil_bool",
			key:   ptr.String("nil_bool"),
			value: nilBool,
			want:  "",
		},
		{
			name:  "nil_int",
			key:   ptr.String("nil_int"),
			value: nilInt,
			want:  "",
		},
		{
			name:  "nil_int64",
			key:   ptr.String("nil_int64"),
			value: nilInt64,
			want:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewProperties()

			p.Set(ptr.ToString(tc.key), tc.value)

			assert.Equal(t, getString(tc.want), p.Get(ptr.ToString(tc.key)))
		})
	}
}

func TestPropertiesSetTag(t *testing.T) {
	cases := []struct {
		name  string
		key   *string
		value interface{}
		want  string
	}{
		{
			name:  "string",
			key:   ptr.String("name"),
			value: "blubber",
			want:  `[tag:name: "blubber"]`,
		},
		{
			name:  "string_ptr",
			key:   ptr.String("name"),
			value: ptr.String("blubber"),
			want:  `[tag:name: "blubber"]`,
		},
		{
			name:  "int",
			key:   ptr.String("int"),
			value: 42,
			want:  `[tag:int: "42"]`,
		},
		{
			name:  "nil",
			key:   ptr.String("nothing"),
			value: nil,
			want:  `[]`,
		},
		{
			name:  "empty_key",
			key:   ptr.String(""),
			value: "empty",
			want:  `[]`,
		},
		{
			name:  "nil_key",
			key:   nil,
			value: "empty",
			want:  `[]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewProperties()

			p.SetTag(tc.key, tc.value)
			have := p.String()

			if tc.want != have {
				t.Errorf("'%s' != '%s'", tc.want, have)
			}
		})
	}
}

func TestPropertiesSetTagWithPrefix(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
		key    *string
		value  interface{}
		want   string
	}{
		{
			name:   "empty",
			prefix: "",
			key:    ptr.String("name"),
			value:  "blubber",
			want:   `[tag:name: "blubber"]`,
		},
		{
			name:   "nonempty",
			prefix: "bish",
			key:    ptr.String("bash"),
			value:  "bosh",
			want:   `[tag:bish:bash: "bosh"]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewProperties()

			p.SetTagWithPrefix(tc.prefix, tc.key, tc.value)
			have := p.String()

			if tc.want != have {
				t.Errorf("'%s' != '%s'", tc.want, have)
			}
		})
	}
}

func TestPropertiesSetPropertiesWithPrefix(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
		key    string
		value  interface{}
		want   string
	}{
		{
			name:   "empty",
			prefix: "",
			key:    "OwnerID",
			value:  ptr.String("123456789012"),
			want:   `[OwnerID: "123456789012"]`,
		},
		{
			name:   "nonempty",
			prefix: "igw",
			key:    "OwnerID",
			value:  ptr.String("123456789012"),
			want:   `[igw:OwnerID: "123456789012"]`,
		},
		{
			name:   "no-property",
			prefix: "igw",
			key:    "",
			value:  ptr.String("123456789012"),
			want:   "[]", // empty properties block
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewProperties()

			p.SetWithPrefix(tc.prefix, tc.key, tc.value)
			have := p.String()

			if tc.want != have {
				t.Errorf("'%s' != '%s'", tc.want, have)
			}
		})
	}
}

func TestPropertiesSetFromStruct(t *testing.T) {
	type testStruct struct {
		Name string
		Age  int
		Tags map[string]string
	}

	type testStruct2 struct {
		Name   string             `property:"name=name"`
		Region *string            `property:"name=region"`
		Tags   *map[string]string `property:"prefix=awesome"`
	}

	type keyValue struct {
		Key   *string
		Value *string
	}

	type tagKeyValue struct {
		TagKey   *string
		TagValue *string
	}

	type customKeyValue struct {
		Name   *string
		Result *string
	}

	type testStruct3 struct {
		Name string `property:""`
		Age  *int   `property:""`
		IQ   *int64 `property:""`
		On   bool
		Off  *bool       `property:"-"`
		Tags []*keyValue `property:""`
	}

	type testStruct3a struct {
		Name string `property:""`
		Age  *int   `property:""`
		IQ   *int64 `property:""`
		On   bool
		Off  *bool          `property:"-"`
		Tags []*tagKeyValue `property:""`
	}

	type testStruct4 struct {
		Omit bool `property:"-"`
		Name byte
	}

	type testStruct5 struct {
		Name string
		Tags map[string]*string
	}

	type testStruct5a struct {
		Name string
		Tags []*customKeyValue `property:"keyField=Name,valueField=Result"`
	}

	type testStruct6 struct {
		Name       string
		Tags       map[*string]*string
		unexported string
	}

	type testStruct7 struct {
		Name   string
		Labels map[string]string `property:"tagPrefix=label"`
	}

	type TestStruct8 struct {
		Name string
	}

	type testStruct9 struct {
		*TestStruct8 `property:",inline"`

		Region string
	}

	type testStruct10 struct {
		*TestStruct8

		Region string
	}

	type testStruct11 struct {
		Name        string
		CreatedTime time.Time
		DeletedTime *time.Time
	}

	type testStruct12 struct {
		Name        string
		CreatedTime time.Time  `property:"prefix=created"`
		DeletedTime *time.Time `property:"prefix=deleted"`
	}

	cases := []struct {
		name  string
		s     interface{}
		want  types.Properties
		error bool
	}{
		{
			name: "empty",
			s:    testStruct{},
			want: types.NewProperties(),
		},
		{
			name: "simple-byte",
			s: testStruct4{
				Name: 'a',
			},
			want: types.NewProperties().Set("Name", "97"),
		},
		{
			name: "from-struct",
			s:    testStruct3{Name: "testing"},
			want: types.NewPropertiesFromStruct(testStruct3{Name: "testing"}),
		},
		{
			name: "simple",
			s:    testStruct{Name: "Alice", Age: 42},
			want: types.NewProperties().Set("Age", 42).Set("Name", "Alice"),
		},
		{
			name: "simple-pointer",
			s:    &testStruct{Name: "Alice", Age: 42},
			want: types.NewProperties().Set("Age", 42).Set("Name", "Alice"),
		},
		{
			name: "complex",
			s: testStruct3{
				Name: "Alice",
				Age:  &[]int{42}[0],
				IQ:   &[]int64{100}[0],
				Off:  &[]bool{true}[0],
				Tags: []*keyValue{
					{Key: ptr.String("key1"), Value: ptr.String("value1")},
				},
			},
			want: types.NewProperties().
				Set("Name", "Alice").
				Set("Age", 42).
				Set("IQ", 100).
				SetTag(ptr.String("key1"), "value1"),
		},
		{
			name: "tags-map",
			s: testStruct2{
				Name:   "Alice",
				Region: ptr.String("us-west-2"),
				Tags:   &map[string]string{"key": "value"},
			},
			want: types.NewProperties().
				Set("name", "Alice").
				Set("region", "us-west-2").
				SetTagWithPrefix("awesome", &[]string{"key"}[0], "value"),
		},
		{
			name: "tags-struct",
			s: testStruct3{
				Name: "Alice",
				Age:  &[]int{42}[0],
				IQ:   &[]int64{100}[0],
				On:   true,
				Tags: []*keyValue{
					{Key: ptr.String("key1"), Value: ptr.String("value1")},
				},
			},
			want: types.NewProperties().
				Set("Name", "Alice").
				Set("Age", 42).
				Set("IQ", 100).
				Set("On", true).
				SetTag(ptr.String("key1"), "value1"),
		},
		{
			name: "tags-struct-tag-key",
			s: testStruct3a{
				Name: "Alice",
				Age:  &[]int{42}[0],
				IQ:   &[]int64{100}[0],
				On:   true,
				Tags: []*tagKeyValue{
					{TagKey: ptr.String("key1"), TagValue: ptr.String("value1")},
				},
			},
			want: types.NewProperties().
				Set("Name", "Alice").
				Set("Age", 42).
				Set("IQ", 100).
				Set("On", true).
				SetTag(ptr.String("key1"), "value1"),
		},
		{
			name: "tags-struct-tag-custom",
			s: testStruct5a{
				Name: "Alice",
				Tags: []*customKeyValue{
					{Name: ptr.String("key1"), Result: ptr.String("value1")},
				},
			},
			want: types.NewProperties().Set("Name", "Alice").SetTag(ptr.String("key1"), "value1"),
		},
		{
			name: "tags-string-pointer",
			s: testStruct5{
				Name: "Alice",
				Tags: map[string]*string{"key": ptr.String("value")},
			},
			want: types.NewProperties().Set("Name", "Alice").SetTag(ptr.String("key"), "value"),
		},
		{
			name: "tags-pointer-pointer",
			s: testStruct6{
				Name:       "Alice",
				Tags:       map[*string]*string{ptr.String("key"): ptr.String("value")},
				unexported: "hidden",
			},
			want: types.NewProperties().Set("Name", "Alice").SetTag(ptr.String("key"), "value"),
		},
		{
			name: "labels-map-string-string",
			s: testStruct7{
				Name:   "Bob",
				Labels: map[string]string{"key": "value"},
			},
			want: types.NewProperties().SetTagPrefix("label").Set("Name", "Bob").SetTag(ptr.String("key"), "value"),
		},
		{
			name: "struct-with-inline",
			s: testStruct9{
				TestStruct8: &TestStruct8{Name: "Alice"},
				Region:      "us-west-2",
			},
			want: types.NewProperties().Set("Name", "Alice").Set("Region", "us-west-2"),
		},
		{
			name: "struct-without-inline",
			s: testStruct10{
				TestStruct8: &TestStruct8{Name: "Alice"},
				Region:      "us-west-2",
			},
			want: types.NewProperties().Set("Region", "us-west-2"),
		},
		{
			name: "time",
			s: testStruct11{
				Name:        "Alice",
				CreatedTime: time.Date(2021, 1, 5, 10, 12, 56, 3309, time.UTC),
				DeletedTime: ptr.Time(time.Date(2023, 7, 15, 5, 32, 12, 4506, time.UTC)),
			},
			want: types.NewProperties().
				Set("Name", "Alice").
				Set("CreatedTime", "2021-01-05T10:12:56Z").
				Set("DeletedTime", "2023-07-15T05:32:12Z"),
		},
		{
			name: "time-with-prefix",
			s: testStruct12{
				Name:        "Alice",
				CreatedTime: time.Date(2021, 1, 5, 10, 12, 56, 3309, time.UTC),
				DeletedTime: ptr.Time(time.Date(2023, 7, 15, 5, 32, 12, 4506, time.UTC)),
			},
			want: types.NewProperties().
				Set("Name", "Alice").
				Set("created:CreatedTime", "2021-01-05T10:12:56Z").
				Set("deleted:DeletedTime", "2023-07-15T05:32:12Z"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewProperties()

			p.SetFromStruct(tc.s)

			assert.Equal(t, tc.want, p)
		})
	}
}

func TestPropertiesSetFromStructNonRepeatableKey(t *testing.T) {
	type testStruct struct {
		Name string
		Age  int
		Tags map[string]string
	}

	type testStruct2 struct {
		ID    string `libnuke:"nonRepeatableKey"`
		State string
	}

	type testStruct3 struct {
		ID    string `property:"name=id" libnuke:"nonRepeatableKey,futureValue"`
		State string `property:"name=state"`
	}

	type testStruct4 struct {
		ID         string `libnuke:"nonRepeatableKey"`
		State      string
		unexported string `libnuke:"nonRepeatableKey"`
	}

	type testStruct5 struct {
		Name         string    `libnuke:"nonRepeatableKey"`
		CreationTime time.Time `libnuke:"nonRepeatableKey"`
		LastEvent    time.Time
	}

	type testStruct6 struct {
		Name         *string    `libnuke:"nonRepeatableKey"`
		CreationTime *time.Time `libnuke:"nonRepeatableKey"`
		LastEvent    *time.Time
	}

	now := time.Now()

	cases := []struct {
		name  string
		s     interface{}
		want  types.Properties
		error bool
	}{
		{
			name: "empty",
			s:    testStruct{},
			want: types.NewProperties(),
		},
		{
			name: "struct-pointer",
			s:    &testStruct{},
			want: types.NewProperties(),
		},
		{
			name: "single-key",
			s: testStruct2{
				ID:    "i-01b489457a60298dd",
				State: "running",
			},
			want: types.NewProperties().Set("ID", "i-01b489457a60298dd"),
		},
		{
			name: "single-key-with-tags",
			s: testStruct3{
				ID:    "i-01b489457a60298dd",
				State: "running",
			},
			want: types.NewProperties().Set("ID", "i-01b489457a60298dd"),
		},
		{
			name: "single-key-with-unexported",
			s: testStruct4{
				ID:         "i-01b489457a60298dd",
				State:      "running",
				unexported: "ignored",
			},
			want: types.NewProperties().Set("ID", "i-01b489457a60298dd"),
		},
		{
			name: "multi-key",
			s: testStruct5{
				Name:         "TestLogGroup",
				CreationTime: now,
				LastEvent:    now,
			},
			want: types.NewProperties().
				Set("Name", "TestLogGroup").
				Set("CreationTime", now),
		},
		{
			name: "multi-key-pointer",
			s: testStruct6{
				Name:         ptr.String("TestLogGroup"),
				CreationTime: &now,
				LastEvent:    &now,
			},
			want: types.NewProperties().
				Set("Name", "TestLogGroup").
				Set("CreationTime", now),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewProperties()

			p.SetFromStructNonRepeatableKey(tc.s)

			assert.Equal(t, tc.want, p)
		})
	}
}

func TestNewNonRepeatableKeyFromStruct(t *testing.T) {
	type testStruct struct {
		ID    string `libnuke:"nonRepeatableKey"`
		State string
	}

	key := types.NewNonRepeatableKeyFromStruct(testStruct{
		ID:    "i-01b489457a60298dd",
		State: "running",
	})

	want := types.NewProperties().Set("ID", "i-01b489457a60298dd")

	assert.Equal(t, want, key)
}

func BenchmarkNewProperties(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = types.NewProperties().
			Set("Name", "Alice").
			Set("Age", 42).
			SetTag(ptr.String("key1"), "value1")
	}
}

func BenchmarkNewPropertiesFromStruct_Simple(b *testing.B) {
	type testStruct struct {
		Name string
		Age  int
	}

	for i := 0; i < b.N; i++ {
		_ = types.NewPropertiesFromStruct(testStruct{Name: "Alice", Age: 42})
	}
}

func BenchmarkNewPropertiesFromStruct_Complex(b *testing.B) {
	type keyValue struct {
		Key   *string
		Value *string
	}

	type testStruct struct {
		Name string
		Age  *int
		Tags []*keyValue
	}

	for i := 0; i < b.N; i++ {
		_ = types.NewPropertiesFromStruct(testStruct{
			Name: "Alice",
			Age:  &[]int{42}[0],
			Tags: []*keyValue{
				{Key: ptr.String("key1"), Value: ptr.String("value1")},
			},
		})
	}
}

func BenchmarkNewPropertiesFromStruct_Complex2(b *testing.B) {
	type keyValue struct {
		TagKey   *string
		TagValue *string
	}

	type testStruct struct {
		Name string
		Age  *int
		Tags []*keyValue
	}

	for i := 0; i < b.N; i++ {
		_ = types.NewPropertiesFromStruct(testStruct{
			Name: "Alice",
			Age:  &[]int{42}[0],
			Tags: []*keyValue{
				{TagKey: ptr.String("key1"), TagValue: ptr.String("value1")},
			},
		})
	}
}

func getString(value interface{}) string {
	switch v := value.(type) {
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case []byte:
		return string(v)
	case *bool:
		if v == nil {
			return ""
		}
		return fmt.Sprint(*v)
	case *int64:
		if v == nil {
			return ""
		}
		return fmt.Sprint(*v)
	case *int:
		if v == nil {
			return ""
		}
		return fmt.Sprint(*v)
	default:
		// Fallback to Stringer interface. This produces gibberish on pointers,
		// but is the only way to avoid reflection.
		return fmt.Sprint(value)
	}
}
