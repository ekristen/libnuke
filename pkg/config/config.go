// Package config provides the configuration for libnuke. It contains the configuration for all the accounts, regions,
// and resource types. It also contains the presets that can be used to apply a set of filters to a nuke process. The
// configuration is loaded from a YAML file and is meant to be used by the implementing tool. Use of the configuration
// is not required but is recommended. The configuration can be implemented a specific way for each tool providing it
// has the necessary methods available.
package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"

	"github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/filter"
)

// Config is the configuration for libnuke. It contains the configuration for all the accounts, regions, and resource
// types. It also contains the presets that can be used to apply a set of filters to a nuke process.
type Config struct {
	Blocklist     []string            `yaml:"blocklist"`
	Regions       []string            `yaml:"regions"`
	Accounts      map[string]*Account `yaml:"accounts"`
	ResourceTypes ResourceTypes       `yaml:"resource-types"`
	Presets       map[string]Preset   `yaml:"presets"`

	AccountBlacklist []string `yaml:"account-blacklist"` // Deprecated: Use Blocklist instead. Will remove in 4.x
	AccountBlocklist []string `yaml:"account-blocklist"` // Deprecated: Use Blocklist instead. Will remove in 4.x

	deprecations map[string]string
	log          *logrus.Entry
}

// Options are the options for creating a new configuration.
type Options struct {
	// Path to the config file
	Path string

	// Logrus entry to use for logging
	Log *logrus.Entry

	// Deprecations is a map of deprecated resource types to their replacements.
	Deprecations map[string]string

	// NoResolveBlacklist will prevent the blocklist from being resolved. This is useful for tools that want to
	// implement their own blocklist.
	NoResolveBlacklist bool

	// NoResolveDeprecations will prevent the deprecations from being resolved. This is useful for tools that want to
	// implement their own deprecations.
	NoResolveDeprecations bool
}

// New creates a new configuration from a file.
func New(opts Options) (*Config, error) {
	c := &Config{
		Accounts:     make(map[string]*Account),
		Presets:      make(map[string]Preset),
		deprecations: make(map[string]string),
	}

	if opts.Log != nil {
		c.log = opts.Log
	} else {
		c.log = logrus.WithField("component", "config")
	}

	if len(opts.Deprecations) > 0 {
		c.deprecations = opts.Deprecations
	}

	err := c.load(opts.Path)
	if err != nil {
		return nil, err
	}

	if !opts.NoResolveBlacklist {
		c.Blocklist = c.ResolveBlocklist()
	}

	if !opts.NoResolveDeprecations {
		if err := c.ResolveDeprecations(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// load loads a configuration from a file and parses it into a Config struct.
func (c *Config) load(path string) error {
	var err error

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(raw, c); err != nil {
		return err
	}

	return nil
}

// ResolveBlocklist returns the blocklist to use to prevent action against the account. In this case account is a
// generic term. It can represent an AWS account, a GCP project, or an Azure tenant.
func (c *Config) ResolveBlocklist() []string {
	var blocklist []string

	if len(c.AccountBlocklist) > 0 {
		blocklist = append(blocklist, c.AccountBlocklist...)
		c.log.Warn("deprecated configuration key 'account-blacklist' - please use 'blocklist' instead")
	}

	if len(c.AccountBlacklist) > 0 {
		blocklist = append(blocklist, c.AccountBlacklist...)
		c.log.Warn("deprecated configuration key 'account-blacklist' - please use 'blocklist' instead")
	}

	if len(c.Blocklist) > 0 {
		blocklist = append(blocklist, c.Blocklist...)
	}

	return blocklist
}

// HasBlocklist returns true if the blocklist is not empty.
func (c *Config) HasBlocklist() bool {
	var blocklist = c.ResolveBlocklist()
	return len(blocklist) > 0
}

// InBlocklist returns true if the searchID is in the blocklist.
func (c *Config) InBlocklist(searchID string) bool {
	for _, blocklistID := range c.ResolveBlocklist() {
		if blocklistID == searchID {
			return true
		}
	}

	return false
}

// ValidateAccount checks the validity of the configuration that's been parsed
func (c *Config) ValidateAccount(accountID string) error {
	if !c.HasBlocklist() {
		return errors.ErrNoBlocklistDefined
	}

	if c.InBlocklist(accountID) {
		return errors.ErrBlocklistAccount
	}

	if _, ok := c.Accounts[accountID]; !ok {
		return errors.ErrAccountNotConfigured
	}

	return nil
}

// Filters resolves all the filters and preset definitions into one set of filters
func (c *Config) Filters(accountID string) (filter.Filters, error) {
	if _, ok := c.Accounts[accountID]; !ok {
		return nil, errors.ErrAccountNotConfigured
	}

	account := c.Accounts[accountID]
	filters := account.Filters

	if filters == nil {
		filters = filter.Filters{}
	}

	if account.Presets == nil {
		return filters, nil
	}

	for _, presetName := range account.Presets {
		notFound := errors.ErrUnknownPreset(presetName)
		if c.Presets == nil {
			return nil, notFound
		}

		preset, ok := c.Presets[presetName]
		if !ok {
			return nil, notFound
		}

		filters.Merge(preset.Filters)
	}

	return filters, nil
}

// ResolveDeprecations resolves any deprecations in the configuration. This is done after the configuration has been
// parsed. It loops through all the accounts and their filters and replaces any deprecated resource types with the
// new resource type.
func (c *Config) ResolveDeprecations() error {
	for _, a := range c.Accounts {
		for resourceType, resources := range a.Filters {
			replacement, ok := c.deprecations[resourceType]
			if !ok {
				continue
			}

			c.log.Warnf("deprecated resource type '%s' - converting to '%s'", resourceType, replacement)
			if _, ok := a.Filters[replacement]; ok {
				return errors.ErrDeprecatedResourceType(
					fmt.Sprintf(
						"using deprecated resource type and replacement: '%s','%s'", resourceType, replacement))
			}

			a.Filters[replacement] = resources

			delete(a.Filters, resourceType)
		}
	}

	return nil
}
