package queue

import (
	"context"
	"fmt"

	"github.com/ekristen/libnuke/pkg/log"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

type ItemState int

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

	iStringer, iOK := i.Resource.(resource.LegacyStringer)
	oStringer, oOK := o.(resource.LegacyStringer)
	if iOK != oOK {
		return false
	}
	if iOK && oOK {
		return iStringer.String() == oStringer.String()
	}

	iGetter, iOK := i.Resource.(resource.PropertyGetter)
	oGetter, oOK := o.(resource.PropertyGetter)
	if iOK != oOK {
		return false
	}
	if iOK && oOK {
		return iGetter.Properties().Equals(oGetter.Properties())
	}

	return false
}

// Print displays the current status of an Item based on it's State
func (i *Item) Print() {
	switch i.State {
	case ItemStateNew:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonWaitPending, "would remove")
	case ItemStateNewDependency:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonWaitDependency, "would remove after dependencies")
	case ItemStateHold:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonHold, "waiting for parent removal")
	case ItemStatePending:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonRemoveTriggered, "triggered remove")
	case ItemStatePendingDependency:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonWaitDependency, fmt.Sprintf("waiting on dependencies (%s)", i.Reason))
	case ItemStateWaiting:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonWaitPending, "waiting")
	case ItemStateFailed:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonError, "failed")
	case ItemStateFiltered:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonSkip, i.Reason)
	case ItemStateFinished:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonSuccess, "removed")
	}
}
