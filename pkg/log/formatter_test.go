package log

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCustomFormatter_Format(t *testing.T) {
	logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

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
			name: "println",
			input: &logrus.Entry{
				Message: "test message",
				Data: logrus.Fields{
					"_handler": "println",
				},
			},
			want: []byte("test message\n"),
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
			name: "missing-name",
			input: &logrus.Entry{
				Data: logrus.Fields{
					"type":       "test",
					"owner":      "owner",
					"state":      "new",
					"state_code": 0,
				},
			},
			want: []byte(`time="0001-01-01T00:00:00Z" level=panic owner=owner state=new state_code=0 type=test
`),
		},
		{
			name: "missing-state-code",
			input: &logrus.Entry{
				Data: logrus.Fields{
					"type":  "test",
					"owner": "owner",
					"state": "new",
					"name":  "resource",
				},
			},
			want: []byte(`time="0001-01-01T00:00:00Z" level=panic name=resource owner=owner state=new type=test
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
					"type":       "test",
					"owner":      "owner",
					"name":       "resource",
					"state":      "new",
					"state_code": 0,
					"prop:one":   "1",
					"prop:two":   "2",
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
					"state":           "hold",
					"state_code":      2,
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
					"type":       "test",
					"owner":      "owner",
					"name":       "resource",
					"state":      "pending",
					"state_code": 3,
					"prop:one":   "1",
					"prop:two":   "2",
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
		name      string
		state     string
		stateCode int
		color     color.Color
	}{
		{
			name:      "reason-success",
			state:     "new",
			stateCode: 0,
			color:     ReasonSuccess,
		},
		{
			name:      "reason-hold",
			state:     "hold",
			stateCode: 2,
			color:     ReasonHold,
		},
		{
			name:      "reason-remove-triggered",
			state:     "pending",
			stateCode: 3,
			color:     ReasonRemoveTriggered,
		},
		{
			name:      "reason-wait-dependency",
			state:     "pending-dependency",
			stateCode: 4,
			color:     ReasonWaitDependency,
		},
		{
			name:      "reason-wait-pending",
			state:     "waiting",
			stateCode: 5,
			color:     ReasonWaitPending,
		},
		{
			name:      "reason-error",
			state:     "failed",
			stateCode: 6,
			color:     ReasonError,
		},
		{
			name:      "reason-skip",
			state:     "filtered",
			stateCode: 7,
			color:     ReasonSkip,
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
			newTestEntry.Data["state_code"] = tc.stateCode

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
