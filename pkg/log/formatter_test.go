package log

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCustomFormatter_Format(t *testing.T) {
	cases := []struct {
		name  string
		input *logrus.Entry
		want  []byte
	}{
		{
			name:  "empty",
			input: nil,
			want:  nil,
		},
		{
			name: "invalid-type",
			input: &logrus.Entry{
				Data: logrus.Fields{
					"type": "test",
				},
			},
			want: []byte(`time="0001-01-01T00:00:00Z" level=panic type=test
`),
		},
		{
			name: "missing-type",
			input: &logrus.Entry{
				Data: logrus.Fields{
					"owner": "owner",
					"name":  "resource",
				},
			},
			want: []byte(`time="0001-01-01T00:00:00Z" level=panic name=resource owner=owner
`),
		},
		{
			name: "missing-owner",
			input: &logrus.Entry{
				Data: logrus.Fields{
					"type": "test",
				},
			},
			want: []byte(`time="0001-01-01T00:00:00Z" level=panic type=test
`),
		},
		{
			name: "missing-resource",
			input: &logrus.Entry{
				Data: logrus.Fields{
					"type":  "test",
					"owner": "owner",
				},
			},
			want: []byte(`time="0001-01-01T00:00:00Z" level=panic owner=owner type=test
`),
		},
		{
			name: "missing-state",
			input: &logrus.Entry{
				Data: logrus.Fields{
					"type":  "test",
					"owner": "owner",
					"name":  "resource",
				},
			},
			want: []byte(`time="0001-01-01T00:00:00Z" level=panic name=resource owner=owner type=test
`),
		},
		{
			name: "reason-success",
			input: &logrus.Entry{
				Message: "would remove",
				Data: logrus.Fields{
					"type":     "test",
					"owner":    "owner",
					"name":     "resource",
					"state":    0,
					"prop:one": "1",
					"prop:two": "2",
				},
			},
			want: []byte(fmt.Sprintf("%s - %s - %s - %s - %s\n",
				ColorRegion.Sprint("owner"),
				ColorResourceType.Sprint("test"),
				ColorResourceID.Sprint("resource"),
				ColorResourceProperties.Sprintf("[%s]", `one: "1"`+", "+`two: "2"`),
				ReasonSuccess.Sprint("would remove"))),
		},
		{
			name: "reason-hold",
			input: &logrus.Entry{
				Message: "test message",
				Data: logrus.Fields{
					"type":            "test",
					"owner":           "owner",
					"name":            "resource",
					"state":           2,
					"prop:one":        "1",
					"prop:two":        "2",
					"prop:_tagPrefix": "tag",
				},
			},
			want: []byte(fmt.Sprintf("%s - %s - %s - %s - %s\n",
				ColorRegion.Sprint("owner"),
				ColorResourceType.Sprint("test"),
				ColorResourceID.Sprint("resource"),
				ColorResourceProperties.Sprintf("[%s]", `one: "1"`+", "+`two: "2"`),
				ReasonHold.Sprint("test message"))),
		},
		{
			name: "reason-remove-triggered",
			input: &logrus.Entry{
				Message: "test message",
				Data: logrus.Fields{
					"type":     "test",
					"owner":    "owner",
					"name":     "resource",
					"state":    3,
					"prop:one": "1",
					"prop:two": "2",
				},
			},
			want: []byte(fmt.Sprintf("%s - %s - %s - %s - %s\n",
				ColorRegion.Sprint("owner"),
				ColorResourceType.Sprint("test"),
				ColorResourceID.Sprint("resource"),
				ColorResourceProperties.Sprintf("[%s]", `one: "1"`+", "+`two: "2"`),
				ReasonRemoveTriggered.Sprint("test message"))),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cf := CustomFormatter{}

			got, err := cf.Format(tc.input)
			assert.NoError(t, err)
			equal := assert.EqualValuesf(t, tc.want, got, "expected %v, got %v", tc.want, got)
			if !equal {
				fmt.Println("`" + string(tc.want) + "`")
				fmt.Println("`" + string(got) + "`")
			}
		})
	}
}

func TestCustomFormatter_FormatReasons(t *testing.T) {
	testEntry := &logrus.Entry{
		Message: "test message",
		Data: logrus.Fields{
			"type":     "test",
			"owner":    "owner",
			"name":     "resource",
			"state":    0,
			"prop:one": "1",
			"prop:two": "2",
		},
	}

	cases := []struct {
		name  string
		state int
		color color.Color
	}{
		{
			name:  "reason-success",
			state: 0,
			color: ReasonSuccess,
		},
		{
			name:  "reason-hold",
			state: 2,
			color: ReasonHold,
		},
		{
			name:  "reason-remove-triggered",
			state: 3,
			color: ReasonRemoveTriggered,
		},
		{
			name:  "reason-wait-dependency",
			state: 4,
			color: ReasonWaitDependency,
		},
		{
			name:  "reason-wait-pending",
			state: 5,
			color: ReasonWaitPending,
		},
		{
			name:  "reason-error",
			state: 6,
			color: ReasonError,
		},
		{
			name:  "reason-skip",
			state: 7,
			color: ReasonSkip,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cf := CustomFormatter{}

			expected := []byte(fmt.Sprintf("%s - %s - %s - %s - %s\n",
				ColorRegion.Sprint("owner"),
				ColorResourceType.Sprint("test"),
				ColorResourceID.Sprint("resource"),
				ColorResourceProperties.Sprintf("[%s]", `one: "1"`+", "+`two: "2"`),
				tc.color.Sprint("test message")))

			newTestEntry := testEntry
			newTestEntry.Data["state"] = tc.state

			got, err := cf.Format(newTestEntry)
			assert.NoError(t, err)
			equal := assert.EqualValues(t, expected, got)
			if !equal {
				t.Errorf("not equal")
				fmt.Println(string(expected))
				fmt.Println(string(got))
			}
		})
	}
}
