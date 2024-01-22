package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"
)

type TestResource struct {
	settings *settings.Setting
}

func (r *TestResource) Remove(_ context.Context) error {
	return fmt.Errorf("remove called")
}

func (r *TestResource) Filter() error {
	return fmt.Errorf("filter called")
}

func (r *TestResource) String() string {
	return "just-a-string"
}

func (r *TestResource) Properties() types.Properties {
	props := types.NewProperties()
	props.Set("test", "example")
	return props
}

func (r *TestResource) Settings(sts *settings.Setting) {
	r.settings = sts
}

func TestInterfaceResource(t *testing.T) {
	r := TestResource{}
	err := r.Remove(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, "remove called", err.Error())
}

func TestInterfaceFilter(t *testing.T) {
	r := TestResource{}
	err := r.Filter()
	assert.Error(t, err)
	assert.Equal(t, "filter called", err.Error())
}

func TestInterfaceLegacyStringer(t *testing.T) {
	r := TestResource{}
	s := r.String()
	assert.Equal(t, "just-a-string", s)
}

func TestInterfacePropertyGetter(t *testing.T) {
	r := TestResource{}
	props := r.Properties()
	assert.Equal(t, "example", props.Get("test"))
}

func TestInterface_SettingsGetter(t *testing.T) {
	s := &settings.Settings{}
	s.Set("TestResource", &settings.Setting{
		"DisableDeletionProtection": true,
	})

	r := TestResource{}
	r.Settings(s.Get("TestResource"))

	assert.Equal(t, true, r.settings.Get("DisableDeletionProtection"))
}
