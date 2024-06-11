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

// HandleWaitHook is an interface that allows a resource to handle waiting for a resource to be deleted.
// This is useful for resources that may take a while to delete, typically where the delete operation happens
// asynchronously from the initial delete command. This allows libnuke to not block during the delete operation.
type HandleWaitHook interface {
	Resource
	HandleWait(context.Context) error
}

// QueueItemHook is an interface that allows a resource to modify the queue item to which it belongs to.
// For advanced use only, please use with caution!
type QueueItemHook interface {
	Resource
	BeforeQueueAdd(interface{})
}
