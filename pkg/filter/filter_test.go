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
			name: "dateOlderThan-invalid-match",
			yaml: `{"type":"dateOlderThan","value":"0"}`,
			match: []string{
				"31-12-2023",
			},
			error: true,
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
	}

	for _, tc := range cases {
		t.Run(tc.yaml, func(t *testing.T) {
			var f filter.Filter

			err := yaml.Unmarshal([]byte(tc.yaml), &f)
			if err != nil {
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
					t.Fatalf("'%v' should match", o)
				}
			}

			for _, o := range tc.mismatch {
				match, err := f.Match(o)
				if err != nil {
					t.Fatal(err)
				}

				if match {
					t.Fatalf("'%v' should not match", o)
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
