package resource

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestLister struct{}

func (l TestLister) List(o interface{}) ([]Resource, error) { return nil, nil }

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
