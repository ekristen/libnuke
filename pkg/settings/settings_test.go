package settings

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
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
	assert.NotNil(t, invalidSettings)

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

func TestSettings_NotNil(t *testing.T) {
	s := Settings{}
	assert.NotNil(t, s.Get("EC2Instance"))
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

func TestSettings_GetNil(t *testing.T) {
	var s *Settings = nil
	set := s.Get("key")
	assert.Nil(t, set)
}

func TestSetting_GetString(t *testing.T) {
	s := Setting{}
	s.Set("TestSetting", "test")
	assert.Equal(t, "test", s.GetString("TestSetting"))
	assert.Equal(t, "", s.GetString("InvalidSetting"))
}

func TestSetting_GetInt(t *testing.T) {
	s := Setting{}
	s.Set("TestSetting", 123)
	assert.Equal(t, 123, s.GetInt("TestSetting"))
	assert.Equal(t, -1, s.GetInt("InvalidSetting"))
}

func TestSetting_GetBool(t *testing.T) {
	s := Setting{}
	s.Set("TestSetting", true)
	assert.Equal(t, true, s.GetBool("TestSetting"))
	assert.Equal(t, false, s.GetBool("InvalidSetting"))
}
