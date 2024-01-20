package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestLister struct{}

func (l TestLister) List(_ context.Context, o interface{}) ([]Resource, error) { return nil, nil }

func Test_RegisterNoScope(t *testing.T) {
	ClearRegistry()

	Register(Registration{
		Name:   "test",
		Lister: TestLister{},
	})

	assert.Len(t, registrations, 1)

	reg := GetRegistration("test")
	assert.Equal(t, DefaultScope, reg.Scope)
	assert.Equal(t, "test", reg.Name)
}

func Test_RegisterResources(t *testing.T) {
	ClearRegistry()

	Register(Registration{
		Name:   "test",
		Scope:  "test",
		Lister: TestLister{},
		DeprecatedAliases: []string{
			"test2",
		},
	})

	if len(registrations) != 1 {
		t.Errorf("expected 1 registration, got %d", len(registrations))
	}

	listers := GetListers()
	assert.Len(t, listers, 1)

	scopeListers := GetListersForScope("test")
	assert.Len(t, scopeListers, 1)

	names := GetNamesForScope("test")
	assert.Len(t, names, 1)

	deprecatedMapping := GetDeprecatedResourceTypeMapping()
	assert.Len(t, deprecatedMapping, 1)
	assert.Equal(t, "test", deprecatedMapping["test2"])
}

func Test_RegisterResourcesDouble(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	Register(Registration{
		Name:   "test",
		Scope:  "test",
		Lister: TestLister{},
	})

	Register(Registration{
		Name:   "test",
		Scope:  "test",
		Lister: TestLister{},
	})
}

func Test_Sorted(t *testing.T) {
	Register(Registration{
		Name:   "Second",
		Scope:  "test",
		Lister: TestLister{},
		DependsOn: []string{
			"First",
		},
	})

	Register(Registration{
		Name:   "First",
		Scope:  "test",
		Lister: TestLister{},
	})

	Register(Registration{
		Name:   "Third",
		Scope:  "test",
		Lister: TestLister{},
		DependsOn: []string{
			"Second",
		},
	})

	Register(Registration{
		Name:   "Fourth",
		Scope:  "test",
		Lister: TestLister{},
		DependsOn: []string{
			"First",
		},
	})

	names := GetNames()
	fmt.Println(names)
}

func Test_RegisterResourcesWithAlternative(t *testing.T) {
	ClearRegistry()

	Register(Registration{
		Name:                "test",
		Scope:               "test",
		Lister:              TestLister{},
		AlternativeResource: "test2",
	})

	Register(Registration{
		Name:   "test2",
		Scope:  "test",
		Lister: TestLister{},
	})

	assert.Len(t, registrations, 2)

	deprecatedMapping := GetAlternativeResourceTypeMapping()
	assert.Len(t, deprecatedMapping, 1)
	assert.Equal(t, "test2", deprecatedMapping["test"])
}
