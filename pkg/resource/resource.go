// Package resource provides a way to interact with resources. This provides multiple interfaces to test against
// as resources can optionally implement these interfaces.
//
// All new resources should implement Properties() and UniqueKey(), but not String(). Without UniqueKey(), libnuke
// attempts to use Properties() to match resources, which does not work if any of the properties change in value during
// the run. UniqueKey() is also useful in older resources that still implement String(), to prevent libnuke from
// mistakenly identifying resources that were recreated with the same String() value after being nuked.
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

// UniqueKeyGetter is an interface that allows a resource to provide a key that uniquely identifies an instance of a
// resource. UniqueKey() can return any field that uniquely identifies the resource and whose value is random or non-
// deterministic. For example, an EC2 instance ID (i-1234567890abcdef0), or resource name with creation time appended.
type UniqueKeyGetter interface {
	Resource
	UniqueKey() string
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
	BeforeEnqueue(interface{})
}
