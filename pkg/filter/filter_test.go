package filter_test

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/filter"
)

func TestNewExactFilter(t *testing.T) {
	f := filter.NewExactFilter("testing")

	assert.Equal(t, f.Type, filter.Exact)

	b1, err := f.Match("testing")
	assert.NoError(t, err)
	assert.True(t, b1)

	b2, err := f.Match("test")
	assert.NoError(t, err)
	assert.False(t, b2)
}

func TestUnmarshalFilter(t *testing.T) {
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

func TestMerge(t *testing.T) {
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

	// Check if the result is as expected
	if !reflect.DeepEqual(f1, expected) {
		t.Errorf("Merge() = %v, want %v", f1, expected)
	}
}
