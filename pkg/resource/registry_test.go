package resource

import (
	"fmt"
	"testing"
)

type TestLister struct{}

func (l TestLister) List(o interface{}) ([]Resource, error) { return nil, nil }

func Test_RegisterResources(t *testing.T) {
	Register(Registration{
		Name:   "test",
		Scope:  "test",
		Lister: TestLister{},
	})

	if len(registrations) != 1 {
		t.Errorf("expected 1 registration, got %d", len(registrations))
	}
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
