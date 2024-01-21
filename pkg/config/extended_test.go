package config

import (
	"bytes"
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type TestCustomService struct {
	Service               string `yaml:"service"`
	URL                   string `yaml:"url"`
	TLSInsecureSkipVerify bool   `yaml:"tls_insecure_skip_verify"`
}

type TestCustomServices []*TestCustomService

// GetService returns the custom region or nil when no such custom endpoints are defined for this region
func (services TestCustomServices) GetService(serviceType string) *TestCustomService {
	for _, s := range services {
		if serviceType == s.Service {
			return s
		}
	}
	return nil
}

type TestCustomRegion struct {
	Region                string             `yaml:"region"`
	Services              TestCustomServices `yaml:"services"`
	TLSInsecureSkipVerify bool               `yaml:"tls_insecure_skip_verify"`
}

type TestCustomEndpoints []*TestCustomRegion

func (endpoints TestCustomEndpoints) GetRegion(region string) *TestCustomRegion {
	for _, r := range endpoints {
		if r.Region == region {
			if r.TLSInsecureSkipVerify {
				for _, s := range r.Services {
					s.TLSInsecureSkipVerify = r.TLSInsecureSkipVerify
				}
			}
			return r
		}
	}
	return nil
}

func (endpoints TestCustomEndpoints) GetURL(region, serviceType string) string {
	r := endpoints.GetRegion(region)
	if r == nil {
		return ""
	}
	s := r.Services.GetService(serviceType)
	if s == nil {
		return ""
	}
	return s.URL
}

type TestExpandedConfig struct {
	*Config         `yaml:",inline"`
	CustomEndpoints TestCustomEndpoints `yaml:"endpoints"`
}

func NewExpandedConfig(opts Options) (*TestExpandedConfig, error) {
	c := &TestExpandedConfig{
		Config: &Config{
			Accounts:     make(map[string]*Account),
			Presets:      make(map[string]Preset),
			deprecations: make(map[string]string),
		},
	}

	if opts.Log != nil {
		c.log = opts.Log
	} else {
		c.log = logrus.NewEntry(logrus.New()).WithField("component", "config")
	}

	err := c.load(opts.Path)
	if err != nil {
		return nil, err
	}

	if !opts.NoResolveBlacklist {
		c.ResolveBlocklist()
	}

	if !opts.NoResolveDeprecations {
		if err := c.ResolveDeprecations(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Load loads a configuration from a file and parses it into a Config struct.
func (c *TestExpandedConfig) load(path string) error {
	var err error

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return err
	}

	return nil
}

func NewTestExpandedConfig(path string) (*TestExpandedConfig, error) {
	c := &TestExpandedConfig{
		Config: &Config{
			Blocklist: make([]string, 0),
			Accounts:  make(map[string]*Account),
			Presets:   make(map[string]Preset),
			Regions:   make([]string, 0),
		},
		CustomEndpoints: make(TestCustomEndpoints, 0),
	}

	if err := c.load(path); err != nil {
		return nil, err
	}

	return c, nil
}

func Test_ExpandedConfig(t *testing.T) {
	expandedCfg, err := NewTestExpandedConfig("testdata/expanded.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "us-east-1", expandedCfg.Regions[0])
	assert.Len(t, expandedCfg.CustomEndpoints, 1)
}
