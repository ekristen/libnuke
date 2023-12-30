package queue

import (
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

type IQueue interface {
	Total() int
	Count(states ...ItemState) int
}

type Queue struct {
	Items []IItem
}
