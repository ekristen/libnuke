package resource

import (
	"testing"
)

func Test_RegisterResources(t *testing.T) {
	Register(Registration{
		Name:   "test",
		Scope:  "test",
		Lister: func(lister ListerOpts) ([]Resource, error) { return nil, nil },
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
		Lister: func(lister ListerOpts) ([]Resource, error) { return nil, nil },
	})

	Register(Registration{
		Name:   "test",
		Scope:  "test",
		Lister: func(lister ListerOpts) ([]Resource, error) { return nil, nil },
	})
}
