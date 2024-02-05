// Package nuke provides the framework for scanning for resources and then iterating over said resources to determine
// if they should be removed or not and in what order.
package nuke

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/sirupsen/logrus"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/scan"
	"github.com/ekristen/libnuke/pkg/types"
	"github.com/ekristen/libnuke/pkg/utils"
)

// ListCache is used to cache the list of resources that are returned from the API.
type ListCache map[string]map[string][]resource.Resource

// Parameters is a collection of common variables used to configure the before of the Nuke instance.
type Parameters struct {
	NoDryRun       bool // NoDryRun instructs Run to actually perform the remove function
	Force          bool // Force instructs Run to proceed without confirmation from user
	ForceSleep     int  // ForceSleep indicates how long of a delay before proceeding with confirmation
	Quiet          bool // Quiet will hide resources if they have been filtered
	MaxWaitRetries int  // MaxWaitRetries is the total number of times a resource will be retried during wait state

	// WaitOnDependencies controls whether resources will be removed after their dependencies. It is important to note
	// that it does not currently track direct dependencies but instead dependent resources. For example if ResourceA
	// depends on ResourceB, all ResourceB has to be in a completed state (removed or failed) before ResourceA will be
	// processed
	WaitOnDependencies bool

	// Includes is a list of resource types that are to be included during the nuke process. If a resource type is
	// listed in both the Includes and Excludes fields then the Excludes field will take precedence.
	Includes []string

	// Excludes is a list of resource types that are to be excluded during the nuke process. If a resource type is
	// listed in both the Includes and Excludes fields then the Excludes field will take precedence.
	Excludes []string

	// Alternatives is a list of resource types that are to be used instead of the default resource. The primary use
	// case for this is AWS Cloud Control API resources.
	Alternatives []string
}

type INuke interface {
	Run() error
	Scan() error
	Filter(item *queue.Item) error
	HandleQueue()
	HandleRemove(item *queue.Item)
	HandleWait(item *queue.Item, cache ListCache)
}

// Nuke is the main struct for the library. It is used to register resource types, scanners, filters and validation
// handlers.
type Nuke struct {
	Parameters *Parameters           // Parameters is a collection of common variables used to configure the before of the Nuke instance.
	Filters    filter.Filters        // Filters is the collection of filters that will be used to filter resources
	Settings   *libsettings.Settings // Settings is the collection of settings that will be used to control resource behavior

	ValidateHandlers []func() error
	ResourceTypes    map[resource.Scope]types.Collection
	Scanners         map[resource.Scope][]*scan.Scanner
	Queue            *queue.Queue // Queue is the queue of resources that will be processed

	scannerHashes []string      // scannerHashes is used to track if a scanner has already been registered
	prompt        func() error  // prompt is what is shown to the user for confirmation
	version       string        // version is what is shown at the beginning of a run
	log           *logrus.Entry // log is the logger that is used for the library
	runSleep      time.Duration // runSleep is how long to sleep between runs of the queue

	failedCount  int // failedCount is used to track how many times we've retried all failed resources
	waitingCount int // waitingCount is used to track how many times we've waiting for resources to move states
}

// New returns an instance of nuke that is properly configured for initial use
func New(params *Parameters, filters filter.Filters, settings *libsettings.Settings) *Nuke {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	n := &Nuke{
		Parameters: params,
		Filters:    filters,
		Queue:      queue.New(),
		Settings:   settings,
		log:        logger.WithField("component", "nuke"),
		runSleep:   5 * time.Second,
	}

	if n.Settings == nil {
		n.Settings = &libsettings.Settings{}
	}

	return n
}

// SetLogger allows the tool instantiating the library to set the logger that is used for the library. It is optional.
func (n *Nuke) SetLogger(logger *logrus.Entry) {
	n.log = logger
}

// SetRunSleep allows the tool instantiating the library to set the sleep duration between runs of the queue.
// It is optional.
func (n *Nuke) SetRunSleep(duration time.Duration) {
	n.runSleep = duration
}

// RegisterVersion allows the tool instantiating the library to register its version so there's consist output
// of the version information across all tools. It is optional.
func (n *Nuke) RegisterVersion(version string) {
	n.version = version
}

// RegisterValidateHandler allows the tool instantiating the library to register a validation handler. It is optional.
func (n *Nuke) RegisterValidateHandler(handler func() error) {
	if n.ValidateHandlers == nil {
		n.ValidateHandlers = make([]func() error, 0)
	}

	n.ValidateHandlers = append(n.ValidateHandlers, handler)
}

// RegisterResourceTypes is used to register resource types against a scope. A scope is a string that is used to
// group resource types together. For example, you could have a scope of "aws" and register all AWS resource types.
// For Azure, you have to register resources by tenant or subscription or even resource group.
func (n *Nuke) RegisterResourceTypes(scope resource.Scope, resourceTypes ...string) {
	if n.ResourceTypes == nil {
		n.ResourceTypes = make(map[resource.Scope]types.Collection)
	}

	n.ResourceTypes[scope] = append(n.ResourceTypes[scope], resourceTypes...)
}

// RegisterScanner is used to register a scanner against a scope. A scope is a string that is used to group resource
// types together. A scanner is what is responsible for actually querying the API for resources and adding them to
// the queue for processing.
func (n *Nuke) RegisterScanner(scope resource.Scope, instance *scan.Scanner) error {
	if n.Scanners == nil {
		n.Scanners = make(map[resource.Scope][]*scan.Scanner)
	}

	hashString := fmt.Sprintf("%s-%s", scope, instance.Owner)
	n.log.Debugf("hash: %s", hashString)
	if slices.Contains(n.scannerHashes, hashString) {
		return fmt.Errorf("scanner is already registered, you cannot register it twice")
	}

	if n.scannerHashes == nil {
		n.scannerHashes = make([]string, 0)
	}

	n.scannerHashes = append(n.scannerHashes, hashString)
	n.Scanners[scope] = append(n.Scanners[scope], instance)

	return nil
}

// RegisterPrompt is used to register the prompt function that used to prompt the user for input, usually to confirm
// if the nuke process should continue or not.
func (n *Nuke) RegisterPrompt(prompt func() error) {
	n.prompt = prompt
}

// Prompt actually calls the registered prompt function as part of the run
func (n *Nuke) Prompt() error {
	if n.prompt != nil {
		return n.prompt()
	}

	return nil
}

// Run is the main entry point for the library. It will run the validation handlers, prompt the user, scan for
// resources, filter them and then process them.
func (n *Nuke) Run(ctx context.Context) error {
	n.Version()

	if err := n.Validate(); err != nil {
		return err
	}

	if err := n.Prompt(); err != nil {
		return err
	}

	if err := n.Scan(ctx); err != nil {
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

	if err := n.Prompt(); err != nil {
		return err
	}

	if err := n.run(ctx); err != nil {
		return err
	}

	fmt.Printf("Nuke complete: %d failed, %d skipped, %d finished.\n\n",
		n.Queue.Count(queue.ItemStateFailed), n.Queue.Count(queue.ItemStateFiltered), n.Queue.Count(queue.ItemStateFinished))

	return nil
}

// handleFailure is used to handle the failure state of resources. It will determine if there have been too many
// failures and exit accordingly, writing to screen the failure state of each resource
func (n *Nuke) handleFailure() error {
	// processingCount is used to determine if there are any resources that are not in the failed state
	processingCount := n.Queue.Count(queue.ItemStatePending, queue.ItemStatePendingDependency, queue.ItemStateHold,
		queue.ItemStateWaiting, queue.ItemStateNew, queue.ItemStateNewDependency)

	// failedCount is used to determine if there are any resources that are in the failed state
	failedCount := n.Queue.Count(queue.ItemStateFailed)

	// if there are no resources being processed and there are resources in the failed state, then we enter this
	// loop to determine how many times we've tried the failed resources
	if processingCount == 0 && failedCount > 0 {
		// if failCount is greater than 2, then we are done, print status and return failed error
		if n.failedCount >= 2 {
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

		n.failedCount++
	} else {
		n.failedCount = 0
	}

	return nil
}

// handleWaiting is used to handle the waiting state of resources. It will determine if there have been too many
// wait retries and exit accordingly.
func (n *Nuke) handleWaiting() error {
	// if MaxWaitRetries is set to 0, then we do not need to do anything, we will retry indefinitely
	if n.Parameters.MaxWaitRetries == 0 {
		return nil
	}

	// pendingCount is used to determine if there are any resources that are still in a pending or hold
	pendingCount := n.Queue.Count(queue.ItemStateWaiting, queue.ItemStatePending,
		queue.ItemStatePendingDependency, queue.ItemStateHold)

	// newCount is used to determine if there are any resources that are still in a new state
	newCount := n.Queue.Count(queue.ItemStateNew, queue.ItemStateNewDependency)

	// If MaxWaitRetries is set, then we need to know if all resources have been moved from new to a pending state.
	// If there are pending, then we need to know how many times to retry before giving up, otherwise we try
	// indefinitely.
	if pendingCount > 0 && newCount == 0 {
		if n.waitingCount >= n.Parameters.MaxWaitRetries {
			return fmt.Errorf("max wait retries of %d exceeded", n.Parameters.MaxWaitRetries)
		}
		n.waitingCount++
	} else {
		n.waitingCount = 0
	}

	return nil
}

// run handles the processing and loop of the queue of items
func (n *Nuke) run(ctx context.Context) error {
	for {
		// HandleQueue is used to handle the queue of resources. It will iterate over the queue and trigger the
		// appropriate handlers based on the state of the resource.
		n.HandleQueue(ctx)

		// handleFailure will check to see if we are in a final failure state and should error out and exit
		if err := n.handleFailure(); err != nil {
			return err
		}

		// handleWaiting will check to see if we have waited to long for resources to retry and error and exit
		if err := n.handleWaiting(); err != nil {
			return err
		}

		// unfinishedCount is used to determine if there are any resources that are still in a state
		// that is not the finished state
		unfinishedCount := n.Queue.Count(queue.ItemStateNew, queue.ItemStateNewDependency,
			queue.ItemStatePending, queue.ItemStatePendingDependency, queue.ItemStateFailed,
			queue.ItemStateWaiting, queue.ItemStateHold,
		)

		// If there are no resources in the queue that are in a state that is not finished, then we are done
		if unfinishedCount == 0 {
			break
		}

		time.Sleep(n.runSleep)
	}

	return nil
}

// Version prints the version that was registered with the library by the invoking tool.
func (n *Nuke) Version() {
	fmt.Println(n.version)
}

// Validate is used to run the validation handlers that were registered with the library by the invoking tool.
func (n *Nuke) Validate() error {
	if n.Parameters.ForceSleep < 3 {
		return fmt.Errorf("value for --force-sleep cannot be less than 3 seconds. This is for your own protection")
	}

	if err := n.Filters.Validate(); err != nil {
		return err
	}

	for _, handler := range n.ValidateHandlers {
		if err := handler(); err != nil {
			return err
		}
	}

	return nil
}

// getScanners is used to condense the scanners down to a single list
func (n *Nuke) getScanners() []*scan.Scanner {
	var allScanners []*scan.Scanner
	for _, scanners := range n.Scanners {
		allScanners = append(allScanners, scanners...)
	}
	return allScanners
}

// runScanner is used to run a scanner and process the items that are returned from the scanner
func (n *Nuke) runScanner(ctx context.Context, scanner *scan.Scanner, itemQueue *queue.Queue) error {
	if err := scanner.Run(ctx); err != nil {
		return err
	}

	for item := range scanner.Items {
		// Experimental Feature
		if n.Parameters.WaitOnDependencies {
			reg := resource.GetRegistration(item.Type)
			if len(reg.DependsOn) > 0 {
				item.State = queue.ItemStateNewDependency
			}
		}

		sGetter, ok := item.Resource.(resource.SettingsGetter)
		if ok {
			sGetter.Settings(n.Settings.Get(item.Type))
		}

		itemQueue.Items = append(itemQueue.Items, item)
		if err := n.Filter(item); err != nil {
			return err
		}

		if item.State == queue.ItemStateFiltered && !n.Parameters.Quiet {
			item.Print()
		}
	}

	return nil
}

// Scan is used to scan for resources. It will run the scanners that were registered with the library by the invoking
// tool. It will also filter the resources based on the filters that were registered. It will also print the current
// status of the resources.
func (n *Nuke) Scan(ctx context.Context) error {
	itemQueue := queue.New()

	scanners := n.getScanners()

	// Iterate over scanners and run them then process their items.
	for _, actualScanner := range scanners {
		if err := n.runScanner(ctx, actualScanner, itemQueue); err != nil {
			return err
		}
	}

	fmt.Printf("Scan complete: %d total, %d nukeable, %d filtered.\n\n",
		itemQueue.Total(), itemQueue.Count(queue.ItemStateNew, queue.ItemStateNewDependency), itemQueue.Count(queue.ItemStateFiltered))

	n.Queue = itemQueue

	return nil
}

// Filter is used to filter resources. It will run the filters that were registered with the instance of Nuke
// and set the state of the resource to filtered if it matches the filter.
func (n *Nuke) Filter(item *queue.Item) error {
	log := n.log.
		WithField("handler", "Filter").
		WithField("type", item.Type)

	if r, ok := item.Resource.(resource.LegacyStringer); ok {
		log = log.WithField("item", r.String())
	}

	checker, ok := item.Resource.(resource.Filter)
	if ok {
		log.Trace("resource had filter function")
		err := checker.Filter()
		if err != nil {
			log.Trace("resource was filtered by resource filter")
			item.State = queue.ItemStateFiltered
			item.Reason = err.Error()

			// Not returning the error, since it could be because of a failed
			// request to the API. We do not want to block the whole nuking,
			// because of an issue on AWS side.
			return nil
		}
	}

	itemFilters, ok := n.Filters[item.Type]
	if !ok {
		log.Tracef("no filters found for type: %s", item.Type)
		return nil
	}

	for _, f := range itemFilters {
		log.
			WithField("prop", f.Property).
			WithField("type", f.Type).
			WithField("value", f.Value).
			Trace("filter details")

		prop, err := item.GetProperty(f.Property)
		if err != nil {
			return err
		}

		log.Tracef("property: %s", prop)

		match, err := f.Match(prop)
		if err != nil {
			return err
		}

		log.Tracef("match: %t", match)

		if utils.IsTrue(f.Invert) {
			log.WithField("orig", match).WithField("new", !match).Trace("filter inverted")
			match = !match
		}

		if match {
			log.Trace("filter matched")
			item.State = queue.ItemStateFiltered
			item.Reason = "filtered by config"
			return nil
		}
	}

	return nil
}

// HandleQueue is used to handle the queue of resources. It will iterate over the queue and trigger the appropriate
// handlers based on the state of the resource.
func (n *Nuke) HandleQueue(ctx context.Context) {
	listCache := make(map[string]map[string][]resource.Resource)

	for _, item := range n.Queue.GetItems() {
		switch item.GetState() {
		case queue.ItemStateNew, queue.ItemStateHold:
			n.HandleRemove(ctx, item)
			item.Print()
		case queue.ItemStateNewDependency, queue.ItemStatePendingDependency:
			n.HandleWaitDependency(ctx, item)
			item.Print()
		case queue.ItemStateFailed:
			n.HandleRemove(ctx, item)
			n.HandleWait(ctx, item, listCache)
			item.Print()
		case queue.ItemStatePending:
			n.HandleWait(ctx, item, listCache)
			item.State = queue.ItemStateWaiting
			item.Print()
		case queue.ItemStateWaiting:
			n.HandleWait(ctx, item, listCache)
			item.Print()
		}
	}

	countWaiting := n.Queue.Count(
		queue.ItemStateWaiting,
		queue.ItemStatePending,
		queue.ItemStatePendingDependency,
		queue.ItemStateNewDependency,
		queue.ItemStateHold,
	)
	countFailed := n.Queue.Count(queue.ItemStateFailed)
	countSkipped := n.Queue.Count(queue.ItemStateFiltered)
	countFinished := n.Queue.Count(queue.ItemStateFinished)

	fmt.Println()
	fmt.Printf("Removal requested: %d waiting, %d failed, %d skipped, %d finished\n\n",
		countWaiting, countFailed, countSkipped, countFinished)
}

// HandleRemove is used to handle the removal of a resource. It will remove the resource and set the state of the
// resource to pending if it was successful or failed if it was not.
func (n *Nuke) HandleRemove(ctx context.Context, item *queue.Item) {
	err := item.Resource.Remove(ctx)
	if err != nil {
		var resErr liberrors.ErrHoldResource
		if errors.As(err, &resErr) {
			item.State = queue.ItemStateHold
			item.Reason = resErr.Error()
			return
		}

		item.State = queue.ItemStateFailed
		item.Reason = err.Error()
		return
	}

	item.State = queue.ItemStatePending
	item.Reason = ""
}

// HandleWaitDependency is used to handle the waiting of a resource. It will check if the resource has any dependencies
// and if it does, it will check if the dependencies have been removed. If they have, it will trigger the remove handler.
func (n *Nuke) HandleWaitDependency(ctx context.Context, item *queue.Item) {
	reg := resource.GetRegistration(item.Type)
	depCount := 0
	for _, dep := range reg.DependsOn {
		cnt := n.Queue.CountByType(dep,
			queue.ItemStateNew, queue.ItemStateNewDependency,
			queue.ItemStatePending, queue.ItemStatePendingDependency,
			queue.ItemStateWaiting, queue.ItemStateHold)
		depCount += cnt
	}

	if depCount == 0 {
		n.HandleRemove(ctx, item)
		return
	}

	item.State = queue.ItemStatePendingDependency
	item.Reason = fmt.Sprintf("left: %d", depCount)
}

// HandleWait is used to handle the waiting of a resource. It will check if the resource has been removed. If it has,
// it will set the state of the resource to finished. If it has not, it will set the state of the resource to waiting.
func (n *Nuke) HandleWait(ctx context.Context, item *queue.Item, cache ListCache) {
	var err error

	ownerID := item.Owner
	_, ok := cache[ownerID]
	if !ok {
		cache[ownerID] = make(map[string][]resource.Resource)
	}

	left, ok := cache[ownerID][item.Type]
	if !ok {
		left, err = item.List(ctx, item.Opts)
		if err != nil {
			item.State = queue.ItemStateFailed
			item.Reason = err.Error()
			return
		}
		cache[ownerID][item.Type] = left
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
