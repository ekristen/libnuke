// Package config provides the configuration for libnuke. It contains the configuration for all the accounts, regions,
// and resource types. It also contains the presets that can be used to apply a set of filters to a nuke process. The
// configuration is loaded from a YAML file and is meant to be used by the implementing tool. Use of the configuration
// is not required but is recommended. The configuration can be implemented a specific way for each tool providing it
// has the necessary methods available.
package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/settings"
)

// Config is the configuration for libnuke. It contains the configuration for all the accounts, regions, and resource
// types. It also contains the presets that can be used to apply a set of filters to a nuke process.
type Config struct {
	// Blocklist is a list of IDs that are to be excluded from the nuke process. In this case account is a generic term.
	// It can represent an AWS account, a GCP project, or an Azure tenant.
	Blocklist []string `yaml:"blocklist"`

	// Regions is a list of regions that are to be included during the nuke process. Region is fairly generic, it can
	// be an AWS region, a GCP region, or an Azure region, or any other region that is supported by the implementing
	// tool.
	Regions []string `yaml:"regions"`

	// Accounts is a map of accounts that are configured a certain way. Account is fairly generic, it can be an AWS
	// account, a GCP project, or an Azure tenant, or any other account that is supported by the implementing tool.
	Accounts map[string]*Account `yaml:"accounts"`

	// ResourceTypes is a collection of resource types that are to be included or excluded from the nuke process.
	ResourceTypes ResourceTypes `yaml:"resource-types"`

	// Presets is a list of presets that are to be used for the configuration. These are global presets that can be used
	// by any account. A Preset can also be defined at the account leve.
	Presets map[string]Preset `yaml:"presets"`

	// Settings is a collection of resource level settings that are to be used by the resource during the nuke process.
	// Resources define their own settings and this allows those settings to be defined in the configuration. The
	// appropriate settings are then passed to the appropriate resource during the nuke process.
	Settings *settings.Settings `yaml:"settings"`

	// Deprecations is a map of deprecated resource types to their replacements. This is passed in as part of the
	// configuration due to the fact the configuration has to resolve the filters in the presets to from any deprecated
	// resource types to their replacements. It cannot be imported from YAML, instead has to be configured post parsing.
	Deprecations map[string]string `yaml:"-"`

	// Log is the logrus entry to use for logging. It cannot be imported from YAML.
	Log *logrus.Entry `yaml:"-"`

	// Deprecated: Use Blocklist instead. Will remove in 4.x
	AccountBlacklist []string `yaml:"account-blacklist"`

	// Deprecated: Use Blocklist instead. Will remove in 4.x
	AccountBlocklist []string `yaml:"account-blocklist"`
}

// Options are the options for creating a new configuration.
type Options struct {
	// Path to the config file
	Path string

	// Log is the logrus entry to use for logging
	Log *logrus.Entry

	// Deprecations is a map of deprecated resource types to their replacements.
	Deprecations map[string]string

	// NoResolveBlacklist will prevent the blocklist from being resolved. This is useful for tools that want to
	// implement their own blocklist. Advanced use only, typically for unit tests.
	NoResolveBlacklist bool

	// NoResolveDeprecations will prevent the Deprecations from being resolved. This is useful for tools that want to
	// implement their own Deprecations. Advanced used only, typically for unit tests.
	NoResolveDeprecations bool
}

// New creates a new configuration from a file.
func New(opts Options) (*Config, error) {
	c := &Config{
		Accounts:     make(map[string]*Account),
		Presets:      make(map[string]Preset),
		Deprecations: make(map[string]string),
		Settings:     &settings.Settings{},
	}

	if opts.Log != nil {
		c.Log = opts.Log
	} else {
		// Create a logger that discards all output
		// The only way output is logged is if the instantiating tool provides a logger
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		c.Log = logger.WithField("component", "config")
	}

	if len(opts.Deprecations) > 0 {
		c.Deprecations = opts.Deprecations
	}

	err := c.Load(opts.Path)
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

// Load loads a configuration from a file and parses it into a Config struct.
func (c *Config) Load(path string) error {
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
		c.AccountBlocklist = nil
		c.Log.Warn("deprecated configuration key 'account-blocklist' - please use 'blocklist' instead")
	}

	if len(c.AccountBlacklist) > 0 {
		blocklist = append(blocklist, c.AccountBlacklist...)
		c.AccountBlacklist = nil
		c.Log.Warn("deprecated configuration key 'account-blacklist' - please use 'blocklist' instead")
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

	if account == nil {
		return nil, errors.ErrAccountNotConfigured
	}

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

		filters.Append(preset.Filters)
	}

	return filters, nil
}

// ResolveDeprecations resolves any Deprecations in the configuration. This is done after the configuration has been
// parsed. It loops through all the accounts and their filters and replaces any deprecated resource types with the
// new resource type.
func (c *Config) ResolveDeprecations() error {
	for _, a := range c.Accounts {
		if a == nil {
			return nil
		}

		// Note: if there are no filters defined, then there's no substitution to perform.
		if a.Filters == nil {
			return nil
		}

		for resourceType, resources := range a.Filters {
			replacement, ok := c.Deprecations[resourceType]
			if !ok {
				continue
			}

			c.Log.Warnf("deprecated resource type '%s' - converting to '%s'", resourceType, replacement)
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
