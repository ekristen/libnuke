package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/resource"
)

type TestLister struct{}

func (l TestLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	return nil, nil
}
func (l TestLister) Close() {}

func Test_RegisterNoScope(t *testing.T) {
	ClearRegistry()

	Register(&Registration{
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

	Register(&Registration{
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

	Register(&Registration{
		Name:   "test",
		Scope:  "test",
		Lister: TestLister{},
	})

	Register(&Registration{
		Name:   "test",
		Scope:  "test",
		Lister: TestLister{},
	})
}

func Test_Sorted(t *testing.T) {
	Register(&Registration{
		Name:   "Second",
		Scope:  "test",
		Lister: TestLister{},
		DependsOn: []string{
			"First",
		},
	})

	Register(&Registration{
		Name:   "First",
		Scope:  "test",
		Lister: TestLister{},
	})

	Register(&Registration{
		Name:   "Third",
		Scope:  "test",
		Lister: TestLister{},
		DependsOn: []string{
			"Second",
		},
	})

	Register(&Registration{
		Name:   "Fourth",
		Scope:  "test",
		Lister: TestLister{},
		DependsOn: []string{
			"First",
		},
	})

	names := GetNames()
	assert.Len(t, names, 5)
}

func Test_RegisterResourcesWithAlternative(t *testing.T) {
	ClearRegistry()

	Register(&Registration{
		Name:                "test",
		Scope:               "test",
		Lister:              TestLister{},
		AlternativeResource: "test2",
	})

	Register(&Registration{
		Name:   "test2",
		Scope:  "test",
		Lister: TestLister{},
	})

	assert.Len(t, registrations, 2)

	deprecatedMapping := GetAlternativeResourceTypeMapping()
	assert.Len(t, deprecatedMapping, 1)
	assert.Equal(t, "test2", deprecatedMapping["test"])
}

func Test_RegisterResourcesWithDuplicateAlternative(t *testing.T) {
	ClearRegistry()

	// Note: this is necessary to test the panic when using coverage and multiple tests
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	Register(&Registration{
		Name:                "test",
		Scope:               "test",
		Lister:              TestLister{},
		AlternativeResource: "test2",
	})

	assert.PanicsWithValue(t, `an alternative resource mapping for test2 already exists`, func() {
		Register(&Registration{
			Name:                "test2",
			Scope:               "test",
			Lister:              TestLister{},
			AlternativeResource: "test2",
		})
	})
}

func Test_GetRegistrations(t *testing.T) {
	ClearRegistry()

	Register(&Registration{
		Name:   "test",
		Scope:  "test",
		Lister: TestLister{},
	})

	regs := GetRegistrations()
	assert.Len(t, regs, 1)
}

func Test_GetLister(t *testing.T) {
	ClearRegistry()

	Register(&Registration{
		Name:   "test",
		Scope:  "test",
		Lister: TestLister{},
	})

	l := GetLister("test")
	assert.NotNil(t, l)
}

func Test_GetListersV2_CircularDependency(t *testing.T) {
	ClearRegistry()

	// Note: this is necessary to test the panic when using coverage and multiple tests
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	Register(&Registration{
		Name:   "A",
		Scope:  "test",
		Lister: TestLister{},
		DependsOn: []string{
			"B",
			"C",
		},
	})

	Register(&Registration{
		Name:   "B",
		Scope:  "test",
		Lister: TestLister{},
		DependsOn: []string{
			"A",
			"C",
		},
	})

	Register(&Registration{
		Name:   "C",
		Scope:  "test",
		Lister: TestLister{},
	})

	assert.Panics(t, func() {
		GetListersV2()
	})
}

func TestExpandNames(t *testing.T) {
	ClearRegistry()

	// Note: this is necessary to test the panic when using coverage and multiple tests
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	rs := []string{"OpsOne", "OpsTwo", "TestingOne", "TestingTwo"}

	for _, r := range rs {
		Register(&Registration{
			Name:   r,
			Scope:  "test",
			Lister: TestLister{},
		})
	}

	cases := []struct {
		name     string
		expected []string
	}{
		{
			name:     "Ops*",
			expected: []string{"OpsOne", "OpsTwo"},
		},
		{
			name:     "OpsOne",
			expected: []string{"OpsOne"},
		},
		{
			name:     "OpsThree",
			expected: []string{"OpsThree"},
		},
		{
			name:     "Ops* Testing*",
			expected: []string{"Ops* Testing*"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expanded := ExpandNames([]string{c.name})

			assert.Equal(t, c.expected, expanded)
		})
	}
}
