package filter_test

import (
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/types"
)

func TestFilter_Nil(t *testing.T) {
	f := filter.Filters{}

	assert.Nil(t, f.Get("resource1"))
}

func TestFilter_Global(t *testing.T) {
	f := filter.Filters{
		filter.Global: []filter.Filter{
			{Property: "prop1", Type: filter.Exact, Value: "value1"},
		},
		"resource1": []filter.Filter{
			{Property: "prop2", Type: filter.Glob, Value: "value2"},
		},
		"resource2": []filter.Filter{
			{Property: "prop3", Type: filter.Regex, Value: "value3"},
		},
	}

	expected := filter.Filters{
		"resource1": []filter.Filter{
			{Property: "prop1", Type: filter.Exact, Value: "value1"},
			{Property: "prop2", Type: filter.Glob, Value: "value2"},
		},
		"resource2": []filter.Filter{
			{Property: "prop1", Type: filter.Exact, Value: "value1"},
			{Property: "prop3", Type: filter.Regex, Value: "value3"},
		},
	}

	assert.Equal(t, expected["resource1"], f.Get("resource1"))
	assert.Equal(t, expected["resource2"], f.Get("resource2"))
}

func TestFilter_GlobalYAML(t *testing.T) {
	data, err := os.ReadFile("testdata/global.yaml")
	assert.NoError(t, err)

	config := struct {
		Filters filter.Filters `yaml:"filters"`
	}{}

	err = yaml.Unmarshal(data, &config)
	assert.NoError(t, err)

	expected := filter.Filters{
		"Resource1": []filter.Filter{
			{Property: "prop3", Type: filter.Exact, Value: "value3", Values: []string{}},
			{Property: "prop1", Type: filter.Exact, Value: "value1", Values: []string{}},
		},
		"Resource2": []filter.Filter{
			{Property: "prop3", Type: filter.Exact, Value: "value3", Values: []string{}},
			{Property: "prop2", Type: filter.Exact, Value: "value2", Values: []string{}},
		},
	}

	assert.Equal(t, expected["Resource1"], config.Filters.Get("Resource1"))
	assert.Equal(t, expected["Resource2"], config.Filters.Get("Resource2"))
}

func TestFilter_UnmarshalYAML_Error(t *testing.T) {
	invalidYAML := `
- invalid
- yaml
`

	var f filter.Filter
	err := yaml.Unmarshal([]byte(invalidYAML), &f)
	assert.Error(t, err, "expected an error when unmarshaling invalid YAML")
}

func TestFilter_GetByGroup(t *testing.T) {
	f := filter.Filters{
		"resource1": []filter.Filter{
			{Property: "prop1", Type: filter.Exact, Value: "value1"},
		},
	}

	rf := f.GetByGroup("resource1")

	for g, filters := range rf {
		assert.Equal(t, "default", g)
		assert.Len(t, filters, 1)
	}

	rf1 := f.GetByGroup("invalidResource1")
	assert.Nil(t, rf1)

	matched, err := f.Match("invalidResourceType", filter.Property(&TestResource{}))
	assert.NoError(t, err)
	assert.False(t, matched)
}

func TestFilter_Match(t *testing.T) {
	cases := []struct {
		name      string
		resource  string
		resources []filter.Property
		filters   []filter.Filter
		filtered  bool
		error     bool
	}{
		{
			name:     "simple-filtered",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{},
			},
			filters: []filter.Filter{
				{Property: "prop1", Type: filter.Exact, Value: "testing"},
			},
			filtered: true,
			error:    false,
		},
		{
			name:     "simple-no-filtered",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{},
			},
			filters: []filter.Filter{
				{Property: "prop1", Type: filter.Exact, Value: "testing1"},
			},
			filtered: false,
			error:    false,
		},
		{
			name:     "simple-no-filtered-invert",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{},
			},
			filters: []filter.Filter{
				{Property: "prop1", Type: filter.Exact, Value: "testing", Invert: true},
			},
			filtered: false,
			error:    false,
		},
		{
			name:     "simple-filtered-invert",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{},
			},
			filters: []filter.Filter{
				{Property: "prop1", Type: filter.Exact, Value: "testing1", Invert: true},
			},
			filtered: true,
			error:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filters := filter.Filters{
				tc.resource: tc.filters,
			}

			for _, r := range tc.resources {
				res, err := filters.Match(tc.resource, r)
				if tc.error {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, tc.filtered, res)
			}
		})
	}
}

func TestFilter_MatchGroup(t *testing.T) {
	cases := []struct {
		name      string
		resource  string
		resources []filter.Property
		filters   filter.Filters
		filtered  bool
		error     bool
	}{
		{
			name:     "single-group-filtered",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{
					Props: types.NewProperties().Set("prop1", "testing"),
				},
			},
			filters: filter.Filters{
				"resource1": []filter.Filter{
					{Property: "prop1", Type: filter.Exact, Value: "testing"},
				},
			},
			filtered: true,
			error:    false,
		},
		{
			name:     "single-group-not-filtered",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{
					Props: types.NewProperties().Set("prop1", "testing"),
				},
			},
			filters: filter.Filters{
				"resource1": []filter.Filter{
					{Property: "prop1", Type: filter.Exact, Value: "testing1"},
				},
			},
			filtered: false,
			error:    false,
		},
		{
			name:     "multiple-group-filtered",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{
					Props: types.NewProperties().Set("prop1", "testing").Set("prop2", "testing2"),
				},
			},
			filters: filter.Filters{
				"resource1": []filter.Filter{
					{Property: "prop1", Type: filter.Exact, Value: "testing", Group: "group1"},
					{Property: "prop2", Type: filter.Exact, Value: "testing2", Group: "group2"},
				},
			},
			filtered: true,
			error:    false,
		},
		{
			name:     "multiple-group-not-filtered",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{
					Props: types.NewProperties().Set("prop1", "testing").Set("prop2", "testing2"),
				},
			},
			filters: filter.Filters{
				"resource1": []filter.Filter{
					{Property: "prop1", Type: filter.Exact, Value: "testing", Group: "group1"},
					{Property: "prop2", Type: filter.Exact, Value: "testing2", Group: "group2"},
					{Property: "prop3", Type: filter.Exact, Value: "testing3", Group: "group3"},
				},
			},
			filtered: true,
			error:    false,
		},
		{
			name:     "single-group-error",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{
					Props: types.NewProperties().Set("prop1", "testing"),
				},
			},
			filters: filter.Filters{
				"resource1": []filter.Filter{
					{Property: "no_stringer", Type: filter.Exact, Value: "testing"},
				},
			},
			filtered: false,
			// TODO: add in log handler checks for error as this throws a warning
			error: false,
		},
		{
			name:     "single-group-invalid-type",
			resource: "resource1",
			resources: []filter.Property{
				&TestResource{
					Props: types.NewProperties().Set("prop1", "testing"),
				},
			},
			filters: filter.Filters{
				"resource1": []filter.Filter{
					{Property: "prop1", Type: "NonExistent", Value: "testing"},
				},
			},
			filtered: false,
			error:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, r := range tc.resources {
				res, err := tc.filters.Match(tc.resource, r)
				if tc.error {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, tc.filtered, res)
			}
		})
	}
}

func TestFilter_NewExactFilter(t *testing.T) {
	f := filter.NewExactFilter("testing")

	assert.Equal(t, f.Type, filter.Exact)

	b1, err := f.Match("testing")
	assert.NoError(t, err)
	assert.True(t, b1)

	b2, err := f.Match("test")
	assert.NoError(t, err)
	assert.False(t, b2)
}

func TestFilter_Validation(t *testing.T) {
	cases := []struct {
		name  string
		yaml  string
		error bool
	}{
		{
			yaml: `{"type":"exact","value":"foo"}`,
		},
		{
			yaml:  `{"type":"exact"}`,
			error: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.yaml, func(t *testing.T) {
			var f filter.Filter

			err := yaml.Unmarshal([]byte(tc.yaml), &f)
			if err != nil {
				t.Fatal(err)
			}

			errValidate := f.Validate()
			if tc.error {
				assert.Error(t, errValidate)
			} else {
				assert.NoError(t, errValidate)
			}
		})
	}
}

func TestFilter_UnmarshalFilter(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour)
	future := time.Now().UTC().Add(24 * time.Hour)
	cases := []struct {
		name            string
		yaml            string
		match, mismatch []string
		error           bool
		yamlError       bool
	}{
		{
			yaml:     `foo`,
			match:    []string{"foo"},
			mismatch: []string{"fo", "fooo", "o", "fo"},
		},
		{
			yaml:     `{"type":"exact","value":"foo"}`,
			match:    []string{"foo"},
			mismatch: []string{"fo", "fooo", "o", "fo"},
		},
		{
			yaml:     `{"type":"glob","value":"b*sh"}`,
			match:    []string{"bish", "bash", "bosh", "bush", "boooooosh", "bsh"},
			mismatch: []string{"woooosh", "fooo", "o", "fo"},
		},
		{
			yaml:     `{"type":"glob","value":"b?sh"}`,
			match:    []string{"bish", "bash", "bosh", "bush"},
			mismatch: []string{"woooosh", "fooo", "o", "fo", "boooooosh", "bsh"},
		},
		{
			yaml:     `{"type":"regex","value":"b[iao]sh"}`,
			match:    []string{"bish", "bash", "bosh"},
			mismatch: []string{"woooosh", "fooo", "o", "fo", "boooooosh", "bsh", "bush"},
		},
		{
			name:  "regex-invalid",
			yaml:  `{"type":"regex","value":"b([iao]sh"}`,
			match: []string{"bish", "bash", "bosh"},
			error: true,
		},
		{
			yaml:     `{"type":"contains","value":"mba"}`,
			match:    []string{"bimbaz", "mba", "bi mba z"},
			mismatch: []string{"bim-baz"},
		},
		{
			yaml: `{"type":"dateOlderThan","value":"0"}`,
			match: []string{strconv.Itoa(int(future.Unix())),
				future.Format("2006-01-02"),
				future.Format("2006/01/02"),
				future.Format("2006-01-02T15:04:05Z"),
				future.Format(time.RFC3339Nano),
				future.Format(time.RFC3339),
			},
			mismatch: []string{"",
				strconv.Itoa(int(past.Unix())),
				past.Format("2006-01-02"),
				past.Format("2006/01/02"),
				past.Format("2006-01-02T15:04:05Z"),
				past.Format(time.RFC3339Nano),
				past.Format(time.RFC3339),
			},
		},
		{
			name: "dateOlderThan-invalid-input",
			yaml: `{"type":"dateOlderThan","value":"-360d4h"}`,
			match: []string{strconv.Itoa(int(future.Unix())),
				future.Format("2006-01-02"),
			},
			error: true,
		},
		{
			name: "dateOlderThan-invalid-filtered",
			yaml: `{"type":"dateOlderThan","value":"0"}`,
			match: []string{
				"31-12-2023",
			},
			error: true,
		},
		{
			name: "dateOlderThanNow",
			yaml: `{"type":"dateOlderThanNow","value":"0"}`,
			match: []string{
				past.Format(time.RFC3339),
			},
			mismatch: []string{
				future.Format(time.RFC3339),
			},
		},
		{
			name: "dateOlderThanNow2",
			yaml: `{"type": "dateOlderThanNow", "value": "-36h"}`, // -36 hours
			match: []string{
				past.Add(-13 * time.Hour).Format(time.RFC3339),
			},
			mismatch: []string{
				past.Format(time.RFC3339),   // -24 hours
				future.Format(time.RFC3339), // +24 hours
			},
		},
		{
			yaml:     `{"type":"prefix","value":"someprefix-"}`,
			match:    []string{"someprefix-1234", "someprefix-someprefix", "someprefix-asdafd"},
			mismatch: []string{"not-someprefix-1234", "not-someprefix-asfda"},
		},
		{
			yaml:     `{"type":"suffix","value":"-somesuffix"}`,
			match:    []string{"12345-somesuffix", "someprefix-somesuffix", "asdfdsa-somesuffix"},
			mismatch: []string{"1235-somesuffix-not", "asdf-not-somesuffix-not"},
		},
		{
			name:  "unknown-filter-type",
			yaml:  `{"type":"custom","value":"does-not-matter"}`,
			match: []string{"12345-somesuffix"},
			error: true,
		},
		{
			name:     "invert-truthy",
			yaml:     `{"type":"exact","value":"foo","invert":"true"}`,
			match:    []string{"foo"},
			mismatch: []string{"bar", "baz"},
		},
		{
			name:      "invert-bad-truthy-value",
			yaml:      `{"type":"exact","value":"foo","invert":"this-is-not-a-bool"}`,
			match:     []string{"foo"},
			mismatch:  []string{"bar", "baz"},
			yamlError: true,
		},
		{
			name:     "invert-true",
			yaml:     `{"type":"exact","value":"foo","invert":true}`,
			match:    []string{"foo"},
			mismatch: []string{"bar", "baz"},
		},
		{
			name:     "not-in",
			yaml:     `{"type":"NotIn","values":["foo","bar"]}`,
			match:    []string{"baz", "qux"},
			mismatch: []string{"foo", "bar"},
		},
		{
			name:     "in",
			yaml:     `{"type":"In","values":["foo","bar"]}`,
			match:    []string{"foo", "bar"},
			mismatch: []string{"baz", "qux"},
		},
		{
			name:     "no-type",
			yaml:     `{"value":"foo"}`,
			match:    []string{"foo"},
			mismatch: []string{"fo", "fooo", "o", "fo"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.yaml, func(t *testing.T) {
			var f filter.Filter

			err := yaml.Unmarshal([]byte(tc.yaml), &f)
			if err != nil {
				if tc.yamlError {
					assert.Error(t, err)
					return
				}

				t.Fatal(err)
			}

			for _, o := range tc.match {
				match, err := f.Match(o)
				if err != nil {
					if tc.error {
						assert.Error(t, err, "received expected error")
						continue
					}

					t.Fatal(err)
				}

				if !match {
					t.Fatalf("'%v' should filtered", o)
				}
			}

			for _, o := range tc.mismatch {
				match, err := f.Match(o)
				if err != nil {
					t.Fatal(err)
				}

				if match {
					t.Fatalf("'%v' should not filtered", o)
				}
			}
		})
	}
}

func TestFilter_Merge(t *testing.T) {
	// Create two Filters objects
	f1 := filter.Filters{
		"resource1": []filter.Filter{
			{Property: "prop1", Type: filter.Exact, Value: "value1"},
		},
	}
	f2 := filter.Filters{
		"resource1": []filter.Filter{
			{Property: "prop2", Type: filter.Glob, Value: "value2"},
		},
		"resource2": []filter.Filter{
			{Property: "prop3", Type: filter.Regex, Value: "value3"},
		},
	}

	// Merge the two Filters objects
	f1.Merge(f2)

	// Create the expected result
	expected := filter.Filters{
		"resource1": []filter.Filter{
			{Property: "prop1", Type: filter.Exact, Value: "value1"},
			{Property: "prop2", Type: filter.Glob, Value: "value2"},
		},
		"resource2": []filter.Filter{
			{Property: "prop3", Type: filter.Regex, Value: "value3"},
		},
	}

	validateErr := expected.Validate()
	assert.NoError(t, validateErr)

	// Check if the result is as expected
	if !reflect.DeepEqual(f1, expected) {
		t.Errorf("Merge() = %v, want %v", f1, expected)
	}
}

func TestFilter_Append(t *testing.T) {
	// Create two Filters objects
	f1 := filter.Filters{
		"resource1": []filter.Filter{
			{Property: "prop1", Type: filter.Exact, Value: "value1"},
		},
	}
	f2 := filter.Filters{
		"resource1": []filter.Filter{
			{Property: "prop2", Type: filter.Glob, Value: "value2"},
		},
		"resource2": []filter.Filter{
			{Property: "prop3", Type: filter.Regex, Value: "value3"},
		},
	}

	// Append the two Filters objects
	f1.Append(f2)

	// Create the expected result
	expected := filter.Filters{
		"resource1": []filter.Filter{
			{Property: "prop1", Type: filter.Exact, Value: "value1"},
			{Property: "prop2", Type: filter.Glob, Value: "value2"},
		},
		"resource2": []filter.Filter{
			{Property: "prop3", Type: filter.Regex, Value: "value3"},
		},
	}

	validateErr := expected.Validate()
	assert.NoError(t, validateErr)

	// Check if the result is as expected
	if !reflect.DeepEqual(f1, expected) {
		t.Errorf("Merge() = %v, want %v", f1, expected)
	}
}

func TestFilter_EmptyType(t *testing.T) {
	f := filter.Filter{
		Property: "Name",
		Type:     "",
		Value:    "anything",
	}

	match, err := f.Match("anything")
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestFilter_ValidateError(t *testing.T) {
	filters := filter.Filters{
		"resource1": []filter.Filter{
			{
				Property: "",
				Type:     filter.Empty,
				Value:    "",
			},
		},
	}
	err := filters.Validate()
	assert.Error(t, err)
}

// TestFilter_Invert tests the parsing of the filter and the invert property as both string and bool
func TestFilter_InvertParsing(t *testing.T) {
	cases := []struct {
		name   string
		yaml   string
		error  bool
		invert bool
	}{
		{
			name:   "invert-true-bool",
			yaml:   `{"type":"exact","value":"foo","invert":true}`,
			invert: true,
		},
		{
			name:   "invert-true-string",
			yaml:   `{"type":"exact","value":"foo","invert":"true"}`,
			invert: true,
		},
		{
			name:   "invert-false-bool",
			yaml:   `{"type":"exact","value":"foo","invert":false}`,
			invert: false,
		},
		{
			name:   "invert-false-string",
			yaml:   `{"type":"exact","value":"foo","invert":"false"}`,
			invert: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.yaml, func(t *testing.T) {
			var f filter.Filter

			err := yaml.Unmarshal([]byte(tc.yaml), &f)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.invert, f.Invert)
		})
	}
}
