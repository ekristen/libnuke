package types_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/gotidy/ptr"

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
