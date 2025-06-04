package unique

import (
	"testing"

	"github.com/gotidy/ptr"
)

func TestFromStruct(t *testing.T) {
	type testStruct struct {
		ID    string `libnuke:"uniqueKey"`
		Name  string `libnuke:"uniqueKey"`
		Other string
	}

	type testStruct2 struct {
		ID    string
		Name  string
		Other string
	}

	type testStruct3 struct {
		ID    string  `libnuke:"uniqueKey"`
		Name  *string `libnuke:"uniqueKey"`
		Other *string
	}

	tests := []struct {
		name   string
		input  interface{}
		unique bool // whether a unique key should be generated
	}{
		{
			name:   "all fields set",
			input:  testStruct{ID: "123", Name: "foo", Other: "bar"},
			unique: true,
		},
		{
			name:   "one field set",
			input:  testStruct{ID: "123", Name: "", Other: "bar"},
			unique: true,
		},
		{
			name:   "no fields set",
			input:  testStruct{ID: "", Name: "", Other: ""},
			unique: true,
		},
		{
			name:   "no unique fields set",
			input:  testStruct2{ID: "matters not", Name: "another test", Other: "bar"},
			unique: false,
		},
		{
			name:   "pointer with unique fields",
			input:  &testStruct3{ID: "123", Name: ptr.String("foo"), Other: ptr.String("bar")},
			unique: true,
		},
		{
			name:   "uniqueKey tag with nil pointer field",
			input:  &testStruct3{ID: "123", Name: nil, Other: nil},
			unique: true,
		},
		{
			name: "uniqueKey tag with empty slice field",
			input: struct {
				IDs []string `libnuke:"uniqueKey"`
			}{IDs: []string{}},
			unique: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := FromStruct(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.unique && (key == nil || *key == "") {
				t.Errorf("expected unique key, got nil or empty string")
			}
			if !tt.unique && key != nil {
				t.Errorf("expected nil, got unique key: %v", *key)
			}
		})
	}

	// Add a new test for a struct with no uniqueKey tags
	t.Run("struct with no uniqueKey tags", func(t *testing.T) {
		type noTagStruct struct {
			ID    string
			Name  string
			Other string
		}
		key, err := FromStruct(noTagStruct{ID: "1", Name: "2", Other: "3"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if key != nil {
			t.Errorf("expected nil for struct with no uniqueKey tags, got: %v", *key)
		}
	})

	t.Run("struct with no uniqueKey tags", func(t *testing.T) {
		type testStructNoTags struct {
			ID    string
			Name  string
			Other string
		}
		key, err := FromStruct(testStructNoTags{ID: "1", Name: "2", Other: "3"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if key != nil {
			t.Errorf("expected nil for struct with no uniqueKey tags, got: %v", *key)
		}
	})

	t.Run("toString covers all types", func(t *testing.T) {
		if toString("foo") != "foo" {
			t.Errorf("toString failed for string")
		}
		if toString(42) == "" {
			t.Errorf("toString failed for int")
		}
		if toString(uint(42)) == "" {
			t.Errorf("toString failed for uint")
		}
		if toString(3.14) == "" {
			t.Errorf("toString failed for float")
		}
		if toString(true) != "true" || toString(false) != "false" {
			t.Errorf("toString failed for bool")
		}
		if toString(struct{}{}) != "" {
			t.Errorf("toString failed for default case")
		}
	})
}

func TestFromStruct_Pointer(t *testing.T) {
	type testStruct struct {
		ID string `libnuke:"uniqueKey"`
	}
	val := &testStruct{ID: "abc"}
	key, err := FromStruct(val)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key == nil || *key == "" {
		t.Error("expected unique key for pointer struct, got nil or empty string")
	}
}

func TestFromStruct_NonStruct(t *testing.T) {
	key, err := FromStruct(123)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != nil {
		t.Errorf("expected nil for non-struct, got: %v", *key)
	}
}

func TestFromStruct_Nil(t *testing.T) {
	key, err := FromStruct(nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != nil {
		t.Errorf("expected nil for nil, got: %v", *key)
	}
}
