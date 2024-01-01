package queue

import (
	"fmt"
	"github.com/ekristen/cloud-nuke-sdk/pkg/log"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
)

type ItemState int

const (
	ItemStateNew ItemState = iota
	ItemStateNewDependency
	ItemStatePending
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

type Item struct {
	Resource resource.Resource
	State    ItemState
	Reason   string
	Type     string
	Owner    string // region/subscription
	Opts     interface{}
}

func (i *Item) GetState() ItemState {
	return i.State
}

func (i *Item) GetReason() string {
	return i.Reason
}

func (i *Item) List(opts interface{}) ([]resource.Resource, error) {
	lister := resource.GetLister(i.Type)
	return lister.List(opts)
}

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

func (i *Item) Print() {
	switch i.State {
	case ItemStateNew:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonWaitPending, "would remove")
	case ItemStateNewDependency:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonWaitDependency, "would remove after dependencies")
	case ItemStatePending:
		log.Log(i.Owner, i.Type, i.Resource, log.ReasonRemoveTriggered, "triggered remove")
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
