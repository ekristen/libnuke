package queue

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/log"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

type ItemState int

func (s ItemState) String() string {
	switch s {
	case ItemStateNew:
		return "new"
	case ItemStateNewDependency:
		return "new-dependency"
	case ItemStateHold:
		return "hold"
	case ItemStatePending:
		return "pending"
	case ItemStatePendingDependency:
		return "pending-dependency"
	case ItemStateWaiting:
		return "waiting"
	case ItemStateFailed:
		return "failed"
	case ItemStateFiltered:
		return "filtered"
	case ItemStateFinished:
		return "finished"
	}
	return "unknown"
}

const (
	ItemStateNew ItemState = iota
	ItemStateNewDependency
	ItemStateHold
	ItemStatePending
	ItemStatePendingDependency
	ItemStateWaiting
	ItemStateFailed
	ItemStateFiltered
	ItemStateFinished
)

type IItem interface {
	resource.Resource
	Print()
	List() ([]resource.Resource, error)
	GetProperty(key string) (string, error)
	Equals(resource.Resource) bool
	GetState() ItemState
}

// Item is used to represent a specific resource, and it's current ItemState in the Queue
type Item struct {
	Resource resource.Resource
	State    ItemState
	Reason   string
	Type     string
	Owner    string // region/subscription
	Opts     interface{}
	Logger   *logrus.Logger
}

// GetState returns the current State of the Item
func (i *Item) GetState() ItemState {
	return i.State
}

// GetReason returns the current Reason for the Item which is usually coupled with an error
func (i *Item) GetReason() string {
	return i.Reason
}

// List calls the List method for the lister for the Type that belongs to the Item which returns
// a list of resources or an error. This primarily is used for the HandleWait function.
func (i *Item) List(ctx context.Context, opts interface{}) ([]resource.Resource, error) {
	return registry.GetLister(i.Type).List(ctx, opts)
}

// GetProperty retrieves the string value of a property on the Item's Resource if it exists.
func (i *Item) GetProperty(key string) (string, error) {
	if key == "" {
		stringer, ok := i.Resource.(resource.LegacyStringer)
		if !ok {
			return "", fmt.Errorf("%T does not support legacy IDs", i.Resource)
		}
		return stringer.String(), nil
	}

	getter, ok := i.Resource.(resource.PropertyGetter)
	if !ok {
		return "", fmt.Errorf("%T does not support custom properties", i.Resource)
	}

	return getter.Properties().Get(key), nil
}

// Equals checks if the current Item is identical to the argument Item in the Queue.
func (i *Item) Equals(o resource.Resource) bool {
	iType := fmt.Sprintf("%T", i.Resource)
	oType := fmt.Sprintf("%T", o)
	if iType != oType {
		return false
	}

	// Compare unique keys if present
	iKeyGetter, iOK := i.Resource.(resource.UniqueKeyGetter)
	oKeyGetter, oOK := o.(resource.UniqueKeyGetter)
	if iOK && oOK {
		return iKeyGetter.UniqueKey() == oKeyGetter.UniqueKey()
	}

	// Fall back to legacy string comparison (may not handle case where resource is recreated during nuke)
	iStringer, iOK := i.Resource.(resource.LegacyStringer)
	oStringer, oOK := o.(resource.LegacyStringer)
	if iOK && oOK {
		return iStringer.String() == oStringer.String()
	}

	// Fall back to property comparison (does not handle case where properties change during nuke)
	iPropertyGetter, iOK := i.Resource.(resource.PropertyGetter)
	oPropertyGetter, oOK := o.(resource.PropertyGetter)
	if iOK && oOK {
		return iPropertyGetter.Properties().Equals(oPropertyGetter.Properties())
	}

	return false
}

// Print displays the current status of an Item based on it's State
func (i *Item) Print() {
	if i.Logger == nil {
		i.Logger = logrus.StandardLogger()
		i.Logger.SetFormatter(&log.CustomFormatter{})
	}

	itemLog := i.Logger.WithFields(logrus.Fields{
		"owner":      i.Owner,
		"type":       i.Type,
		"state":      i.State.String(),
		"state_code": int(i.State),
	})

	rString, ok := i.Resource.(resource.LegacyStringer)
	if ok {
		itemLog = itemLog.WithField("name", rString.String())
	}

	rProp, ok := i.Resource.(resource.PropertyGetter)
	if ok {
		itemLog = itemLog.WithFields(sorted(rProp.Properties()))
	}

	switch i.State {
	case ItemStateNew:
		itemLog.Info("would remove")
	case ItemStateNewDependency:
		itemLog.Info("would remove after dependencies")
	case ItemStateHold:
		itemLog.Info("waiting for parent removal")
	case ItemStatePending:
		itemLog.Info("triggered remove")
	case ItemStatePendingDependency:
		itemLog.Infof("waiting on dependencies (%s)", i.Reason)
	case ItemStateWaiting:
		itemLog.Info("waiting for removal")
	case ItemStateFailed:
		itemLog.Info("failed")
	case ItemStateFiltered:
		itemLog.Infof("filtered: %s", i.Reason)
	case ItemStateFinished:
		itemLog.Info("removed")
	}
}

// sorted -- Format the resource properties in sorted order ready for printing.
// This ensures that multiple runs of aws-nuke produce stable output so
// that they can be compared with each other.
func sorted(m map[string]string) logrus.Fields {
	out := logrus.Fields{}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for k := range keys {
		if strings.HasPrefix(keys[k], "_") {
			continue
		}

		out[fmt.Sprintf("prop:%s", keys[k])] = m[keys[k]]
	}
	return out
}
