package settings

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

type Config struct {
	Settings Settings `yaml:"settings"`
}

func TestSettings_ParseYAML(t *testing.T) {
	var cfg Config

	data, err := os.ReadFile("testdata/settings.yaml")
	assert.NoError(t, err)

	err = yaml.Unmarshal(data, &cfg)
	assert.NoError(t, err)

	ec2Settings := cfg.Settings.Get("EC2Instance")
	assert.NotNil(t, ec2Settings)

	assert.Equal(t, true, ec2Settings.Get("DisableDeletionProtection"))
	assert.Equal(t, "true", ec2Settings.Get("DisableStopProtection"))
	assert.Nil(t, ec2Settings.Get("ForceDeleteLightsailAddOns"))

	invalidSettings := cfg.Settings.Get("OtherInstance")
	assert.Nil(t, invalidSettings)

	typeSettings := cfg.Settings.Get("Types")
	assert.NotNil(t, typeSettings)
	assert.Equal(t, 1, typeSettings.Get("Integer"))
	assert.Equal(t, "string", typeSettings.Get("String"))
	assert.NotNil(t, typeSettings.Get("Nested"))
}

func TestSettings_ParseYAMLInvalid(t *testing.T) {
	var cfg Config

	data, err := os.ReadFile("testdata/settings-invalid.yaml")
	assert.NoError(t, err)

	err = yaml.Unmarshal(data, &cfg)
	assert.Error(t, err)
}

func TestSettings_Nil(t *testing.T) {
	s := Settings{}
	assert.Nil(t, s.Get("EC2Instance"))
}

func TestSettings_Set(t *testing.T) {
	s := Settings{}
	s.Set("EC2Instance", &Setting{
		"DisableDeletionProtection": true,
	})
	s.Set("EC2Instance", &Setting{
		"DisableStopProtection": true,
	})
	assert.Equal(t, true, s.Get("EC2Instance").Get("DisableDeletionProtection"))
	assert.Equal(t, true, s.Get("EC2Instance").Get("DisableStopProtection"))
	assert.Nil(t, s.Get("EC2Instance").Get("ForceDeleteLightsailAddOns"))
}

func TestSettings_SetSetting(t *testing.T) {
	s := Setting{}
	s.Set("DisableDeletionProtection", true)
	assert.Equal(t, true, s.Get("DisableDeletionProtection"))
	assert.Nil(t, s.Get("DisableStopProtection"))
	assert.Nil(t, s.Get("ForceDeleteLightsailAddOns"))
}
