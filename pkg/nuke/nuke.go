package nuke

import (
	"fmt"
	"github.com/ekristen/cloud-nuke-sdk/pkg/config"
	"github.com/ekristen/cloud-nuke-sdk/pkg/queue"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
	"github.com/ekristen/cloud-nuke-sdk/pkg/utils"
)

type ListCache map[string]map[string][]resource.Resource

type Parameters struct {
	ConfigPath string

	ID string

	Targets  []string
	Excludes []string

	NoDryRun   bool
	Force      bool
	ForceSleep int
	Quiet      bool

	MaxWaitRetries int
}

type INuke interface {
	Run() error
	Scan() error
	Filter(item *queue.Item) error
	HandleQueue()
	HandleRemove(item *queue.Item)
	HandleWait(item *queue.Item, cache ListCache)
}

type Nuke struct {
	Parameters    Parameters
	Config        config.IConfig
	ResourceTypes types.Collection
	Queue         queue.Queue
	scopes        []resource.Scope
}

func (n *Nuke) Run() error {
	return nil
}

func (n *Nuke) Scan() error {
	return nil
}

func (n *Nuke) Filter(item *queue.Item) error {
	checker, ok := item.Resource.(resource.Filter)
	if ok {
		err := checker.Filter()
		if err != nil {
			item.State = queue.ItemStateFiltered
			item.Reason = err.Error()

			// Not returning the error, since it could be because of a failed
			// request to the API. We do not want to block the whole nuking,
			// because of an issue on AWS side.
			return nil
		}
	}

	accountFilters, err := n.Config.Filters(n.Parameters.ID)
	if err != nil {
		return err
	}

	itemFilters, ok := accountFilters[item.Type]
	if !ok {
		return nil
	}

	for _, filter := range itemFilters {
		prop, err := item.GetProperty(filter.Property)
		if err != nil {
			return err
		}

		match, err := filter.Match(prop)
		if err != nil {
			return err
		}

		if utils.IsTrue(filter.Invert) {
			match = !match
		}

		if match {
			item.State = queue.ItemStateFiltered
			item.Reason = "filtered by config"
			return nil
		}
	}

	return nil
}

func (n *Nuke) HandleQueue() {
	listCache := make(map[string]map[string][]resource.Resource)

	for _, item := range n.queue.GetItems() {
		switch item.GetState() {
		case queue.ItemStateNew:
			n.HandleRemove(item)
			item.Print()
		case queue.ItemStateFailed:
			n.HandleRemove(item)
			n.HandleWait(item, listCache)
			item.Print()
		case queue.ItemStatePending:
			n.HandleWait(item, listCache)
			item.State = queue.ItemStateWaiting
			item.Print()
		case queue.ItemStateWaiting:
			n.HandleWait(item, listCache)
			item.Print()
		}

	}

	fmt.Println()
	fmt.Printf("Removal requested: %d waiting, %d failed, %d skipped, %d finished\n\n",
		n.queue.Count(queue.ItemStateWaiting, queue.ItemStatePending), n.queue.Count(queue.ItemStateFailed),
		n.queue.Count(queue.ItemStateFiltered), n.queue.Count(queue.ItemStateFinished))
}

func (n *Nuke) HandleRemove(item *queue.Item) {
	err := item.Resource.Remove()
	if err != nil {
		item.State = queue.ItemStateFailed
		item.Reason = err.Error()
		return
	}

	item.State = queue.ItemStatePending
	item.Reason = ""
}

func (n *Nuke) HandleWait(item *queue.Item, cache ListCache) {
	var err error
	ownerId := item.Owner
	_, ok := cache[ownerId]
	if !ok {
		cache[ownerId] = make(map[string][]resource.Resource)
	}
	left, ok := cache[ownerId][item.Type]
	if !ok {
		left, err = item.List()
		if err != nil {
			item.State = queue.ItemStateFailed
			item.Reason = err.Error()
			return
		}
		cache[ownerId][item.Type] = left
	}

	for _, r := range left {
		if item.Equals(r) {
			checker, ok := r.(resource.Filter)
			if ok {
				err := checker.Filter()
				if err != nil {
					break
				}
			}

			return
		}
	}

	item.State = queue.ItemStateFinished
	item.Reason = ""
}
