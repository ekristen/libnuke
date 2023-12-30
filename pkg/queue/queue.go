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

type Item interface {
	resource.Resource
	Print()
	List() ([]resource.Resource, error)
	GetProperty(key string) (string, error)
	Equals(resource.Resource) bool
}

type Queue interface {
	Total() int
	Count(states ...ItemState) int
}
