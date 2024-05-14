// Package resource provides a way to interact with resources. This provides multiple interfaces to test against
// as resources can optionally implement these interfaces.
package resource

import (
	"context"

	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"
)

type Resource interface {
	Remove(ctx context.Context) error
}

type Filter interface {
	Resource
	Filter() error
}

type LegacyStringer interface {
	Resource
	String() string
}

type PropertyGetter interface {
	Resource
	Properties() types.Properties
}

type SettingsGetter interface {
	Resource
	Settings(setting *settings.Setting)
}

type HandleWaitHook interface {
	Resource
	HandleWait(context.Context) error
}
