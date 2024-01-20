package config

import (
	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/filter"
)

// IConfig is the interface for the config package. It is used to define the methods that are required for the
// configuration to be used by libnuke. If you are implementing a tool that uses libnuke then you will need to implement
// this interface for your configuration or use the build in config package.
type IConfig interface {
	SetLog(log *logrus.Entry)
	ResolveBlocklist() []string
	HasBlocklist() bool
	InBlocklist(searchID string) bool
	Validate(accountID string) error
	Filters(accountID string) (filter.Filters, error)
	SetDeprecations(deprecations map[string]string)
	ResolveDeprecations() error
}
