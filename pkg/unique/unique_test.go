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
			key := FromStruct(tt.input)
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
		key := FromStruct(noTagStruct{ID: "1", Name: "2", Other: "3"})
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
		key := FromStruct(testStructNoTags{ID: "1", Name: "2", Other: "3"})
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
	key := FromStruct(val)
	if key == nil || *key == "" {
		t.Error("expected unique key for pointer struct, got nil or empty string")
	}
}

func TestFromStruct_NonStruct(t *testing.T) {
	key := FromStruct(123)
	if key != nil {
		t.Errorf("expected nil for non-struct, got: %v", *key)
	}
}

func TestFromStruct_Nil(t *testing.T) {
	key := FromStruct(nil)
	if key != nil {
		t.Errorf("expected nil for nil, got: %v", *key)
	}
}

func TestGenerate_HashDeterminism(t *testing.T) {
	tests := []struct {
		name   string
		input  []interface{}
		expect string
	}{
		{
			name:   "simple strings",
			input:  []interface{}{"foo", "bar"},
			expect: "e3066f35bf392a7f15f2ec7497bcafd3330d83b9f83a2df979b65df9d3bdeef9", // placeholder, replace with actual
		},
		{
			name:   "numbers and bool",
			input:  []interface{}{123, true, 45.6},
			expect: "55f056c6cecba8e5b99ebae70af5e4ff1c476e1c35f3ee8d161b29de4468d459", // placeholder
		},
		{
			name:   "struct",
			input:  []interface{}{struct{ A string }{A: "x"}},
			expect: "cfd8d2a6c2fa8b7693ab816dc210246ff285a42bf21124ece9f869326af35e24", // placeholder
		},
		{
			name:   "slice",
			input:  []interface{}{[]int{1, 2, 3}},
			expect: "a615eeaee21de5179de080de8c3052c8da901138406ba71c38c032845f7d54f4", // placeholder
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := Generate(tt.input...)
			if hash != tt.expect {
				t.Errorf("expected %s, got %s", tt.expect, hash)
			}
		})
	}
}
