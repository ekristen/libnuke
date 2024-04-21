package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateProperties(t *testing.T) {
	type TestResource1 struct {
		Name   string            `description:"The name of the resource"`
		Region *string           `description:"The region in which the resource resides"`
		VpcID  string            `description:"The VPC ID of the resource" property:"prefix=vpc"`
		Tags   map[string]string `description:"The tags associated with the resource"`
	}

	type TestResource2 struct {
		Name   string            `description:"The name of the resource"`
		Region *string           `description:"The region in which the resource resides"`
		Tags   map[string]string `description:"The tags associated with the resource" property:"prefix=ee"`
	}

	type TestResource3 struct {
		Name    string `description:"The name of the resource"`
		Ignore  string `property:"-"`
		Example string `description:"A property rename" property:"name=Delta"`
		skipped string
	}

	cases := []struct {
		name string
		in   interface{}
		want map[string]string
	}{
		{
			name: "TestResource1",
			in:   TestResource1{},
			want: map[string]string{
				"Name":      "The name of the resource",
				"Region":    "The region in which the resource resides",
				"vpc:VpcID": "The VPC ID of the resource",
				"tag:<key>:": "This resource has tags with property `Tags`. These are key/value pairs that are\n\t" +
					"added as their own property with the prefix of `tag:` (e.g. [tag:example: \"value\"]) ",
			},
		},
		{
			name: "TestResource2",
			in:   TestResource2{},
			want: map[string]string{
				"Name":   "The name of the resource",
				"Region": "The region in which the resource resides",
				"tag:ee:<key>:": "This resource has tags with property `Tags`. These are key/value pairs that are\n\t" +
					"added as their own property with the prefix of `tag:ee:" +
					"` (e.g. [tag:ee:example: \"value\"]) ",
			},
		},
		{
			name: "TestResource3",
			in:   TestResource3{},
			want: map[string]string{
				"Name":  "The name of the resource",
				"Delta": "A property rename",
			},
		},
		{
			name: "PointerTestResource3",
			in:   &TestResource3{},
			want: map[string]string{
				"Name":  "The name of the resource",
				"Delta": "A property rename",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			have := GeneratePropertiesMap(c.in)
			assert.Equal(t, c.want, have)
		})
	}
}
