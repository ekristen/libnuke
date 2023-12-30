package queue

import (
	"fmt"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
)

type ItemState int

const (
	ItemStateNew ItemState = iota
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
}

type Item struct {
	Resource resource.Resource
	State    ItemState
	Reason   string
}

func (i *Item) List() ([]resource.Resource, error) {
	return nil, nil
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

func (i *Item) Print() {}
