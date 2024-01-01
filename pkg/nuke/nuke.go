package nuke

import (
	"fmt"
	"github.com/ekristen/cloud-nuke-sdk/pkg/config"
	"github.com/ekristen/cloud-nuke-sdk/pkg/queue"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
	"github.com/ekristen/cloud-nuke-sdk/pkg/utils"
	"github.com/sirupsen/logrus"
	"time"
)

type ListCache map[string]map[string][]resource.Resource

type Parameters struct {
	ConfigPath string

	ID string

	Targets      []string
	Excludes     []string
	CloudControl []string

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
	Parameters Parameters
	Config     config.IConfig
	Queue      queue.Queue
	scopes     []resource.Scope

	ValidateHandlers []func() error

	ResourceTypes map[resource.Scope]types.Collection
	Scanners      map[resource.Scope]*Scanner

	prompts map[string]func() error
}

func (n *Nuke) RegisterValidateHandler(handler func() error) {
	if n.ValidateHandlers == nil {
		n.ValidateHandlers = make([]func() error, 0)
	}

	n.ValidateHandlers = append(n.ValidateHandlers, handler)
}

func (n *Nuke) RegisterResourceTypes(scope resource.Scope, resourceTypes ...string) {
	if n.ResourceTypes == nil {
		n.ResourceTypes = make(map[resource.Scope]types.Collection)
	}

	n.ResourceTypes[scope] = append(n.ResourceTypes[scope], resourceTypes...)
}

func (n *Nuke) RegisterScanner(scope resource.Scope, scanner *Scanner) {
	if n.Scanners == nil {
		n.Scanners = make(map[resource.Scope]*Scanner)
	}

	n.Scanners[scope] = scanner
}

func (n *Nuke) RegisterPrompt(name string, prompt func() error) {
	if n.prompts == nil {
		n.prompts = make(map[string]func() error)
	}

	n.prompts[name] = prompt
}

func (n *Nuke) PromptFirst() error {
	if prompt, ok := n.prompts["first"]; ok {
		return prompt()
	}

	return nil
}

func (n *Nuke) PromptSecond() error {
	if prompt, ok := n.prompts["second"]; ok {
		return prompt()
	}

	return nil
}

func (n *Nuke) Run() error {
	if err := n.Validate(); err != nil {
		return err
	}

	if err := n.PromptFirst(); err != nil {
		return err
	}

	if err := n.Scan(); err != nil {
		return err
	}

	if n.Queue.Count(queue.ItemStateNew) == 0 {
		fmt.Println("No resource to delete.")
		return nil
	}

	if !n.Parameters.NoDryRun {
		fmt.Println("The above resources would be deleted with the supplied configuration. Provide --no-dry-run to actually destroy resources.")
		return nil
	}

	if err := n.PromptSecond(); err != nil {
		return err
	}

	failCount := 0
	waitingCount := 0

	for {
		n.HandleQueue()

		if n.Queue.Count(queue.ItemStatePending, queue.ItemStateWaiting, queue.ItemStateNew, queue.ItemStateNewDependency) == 0 && n.Queue.Count(queue.ItemStateFailed) > 0 {
			if failCount >= 2 {
				logrus.Errorf("There are resources in failed state, but none are ready for deletion, anymore.")
				fmt.Println()

				for _, item := range n.Queue.GetItems() {
					if item.GetState() != queue.ItemStateFailed {
						continue
					}

					item.Print()
					logrus.Error(item.GetReason())
				}

				return fmt.Errorf("failed")
			}

			failCount = failCount + 1
		} else {
			failCount = 0
		}
		if n.Parameters.MaxWaitRetries != 0 && n.Queue.Count(queue.ItemStateWaiting, queue.ItemStatePending) > 0 && n.Queue.Count(queue.ItemStateNew, queue.ItemStateNewDependency) == 0 {
			if waitingCount >= n.Parameters.MaxWaitRetries {
				return fmt.Errorf("Max wait retries of %d exceeded.\n\n", n.Parameters.MaxWaitRetries)
			}
			waitingCount = waitingCount + 1
		} else {
			waitingCount = 0
		}
		if n.Queue.Count(queue.ItemStateNew, queue.ItemStateNewDependency, queue.ItemStatePending, queue.ItemStateFailed, queue.ItemStateWaiting) == 0 {
			break
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Printf("Nuke complete: %d failed, %d skipped, %d finished.\n\n",
		n.Queue.Count(queue.ItemStateFailed), n.Queue.Count(queue.ItemStateFiltered), n.Queue.Count(queue.ItemStateFinished))

	return nil
}

func (n *Nuke) Version() {

}

func (n *Nuke) Validate() error {
	if n.Parameters.ForceSleep < 3 {
		return fmt.Errorf("value for --force-sleep cannot be less than 3 seconds. This is for your own protection")
	}

	n.Version()

	for _, handler := range n.ValidateHandlers {
		if err := handler(); err != nil {
			return err
		}
	}

	return nil
}

func (n *Nuke) Scan() error {
	itemQueue := queue.Queue{
		Items: make([]*queue.Item, 0),
	}

	for _, scanner := range n.Scanners {
		scanner.Run()
		for item := range scanner.Items {
			itemQueue.Items = append(itemQueue.Items, item)
			err := n.Filter(item)
			if err != nil {
				return err
			}

			if item.State != queue.ItemStateFiltered || !n.Parameters.Quiet {
				item.Print()
			}
		}
	}

	fmt.Printf("Scan complete: %d total, %d nukeable, %d filtered.\n\n",
		itemQueue.Count(), itemQueue.Count(queue.ItemStateNew), itemQueue.Count(queue.ItemStateFiltered))

	n.Queue = itemQueue

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

	for _, item := range n.Queue.GetItems() {
		switch item.GetState() {
		case queue.ItemStateNew:
			n.HandleRemove(item)
			item.Print()
		case queue.ItemStateNewDependency:
			n.HandleWaitDependency(item)
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
		n.Queue.Count(queue.ItemStateWaiting, queue.ItemStatePending, queue.ItemStateNewDependency), n.Queue.Count(queue.ItemStateFailed),
		n.Queue.Count(queue.ItemStateFiltered), n.Queue.Count(queue.ItemStateFinished))
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

func (n *Nuke) HandleWaitDependency(item *queue.Item) {
	reg := resource.GetRegistration(item.Type)
	depCount := 0
	for _, dep := range reg.DependsOn {
		cnt := n.Queue.CountByType(dep, queue.ItemStateNew, queue.ItemStatePending, queue.ItemStateWaiting)
		depCount = depCount + cnt
	}

	if depCount == 0 {
		n.HandleRemove(item)
	}
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
		left, err = item.List(item.Opts)
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
