package nuke

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	sdkerrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/utils"
)

// ScannerParallelQueries is the number of parallel queries to run at any given time for a scanner.
const ScannerParallelQueries = 16

// Scanner is collection of resource types that will be scanned for existing resources and added to the
// item queue for processing. These items will be filtered and then processed.
type Scanner struct {
	Items     chan *queue.Item
	semaphore *semaphore.Weighted

	resourceTypes  []string
	options        interface{}
	owner          string
	mutateOptsFunc MutateOptsFunc
}

// MutateOptsFunc is a function that can mutate the options for a given resource type. This is useful for when you
// need to pass in a different set of options for a given resource type. For example, AWS nuke needs to be able to
// populate the region and session for a given resource type give that it might only exist in us-east-1.
type MutateOptsFunc func(opts interface{}, resourceType string) interface{}

// NewScanner creates a new scanner for the given resource types.
func NewScanner(owner string, resourceTypes []string, opts interface{}) *Scanner {
	return &Scanner{
		Items:         make(chan *queue.Item, 10000),
		semaphore:     semaphore.NewWeighted(ScannerParallelQueries),
		resourceTypes: resourceTypes,
		options:       opts,
		owner:         owner,
	}
}

type IScanner interface {
	Run(resourceTypes []string)
	list(resourceType string)
}

// RegisterMutateOptsFunc registers a mutate options function for the scanner. The mutate options function is called
// for each resource type that is being scanned. This allows you to mutate the options for a given resource type.
func (s *Scanner) RegisterMutateOptsFunc(morph MutateOptsFunc) {
	if s.mutateOptsFunc != nil {
		panic("mutateOptsFunc already registered")
	}

	s.mutateOptsFunc = morph
}

// Run starts the scanner and runs the lister for each resource type.
func (s *Scanner) Run() error {
	ctx := context.Background()

	for _, resourceType := range s.resourceTypes {
		err := s.semaphore.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		opts := s.options
		if s.mutateOptsFunc != nil {
			opts = s.mutateOptsFunc(opts, resourceType)
		}

		go s.list(s.owner, resourceType, opts)
	}

	// Wait for all routines to finish.
	err := s.semaphore.Acquire(ctx, ScannerParallelQueries)
	if err != nil {
		return err
	}

	close(s.Items)

	return nil
}

func (s *Scanner) list(owner, resourceType string, opts interface{}) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v\n\n%s", r.(error), string(debug.Stack()))
			dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
			logrus.Errorf("Listing %s failed:\n%s", resourceType, dump)
		}
	}()
	defer s.semaphore.Release(1)

	lister := resource.GetLister(resourceType)
	var rs []resource.Resource

	rs, err := lister.List(opts)
	if err != nil {
		var errSkipRequest sdkerrors.ErrSkipRequest
		ok := errors.As(err, &errSkipRequest)
		if ok {
			logrus.Debugf("skipping request: %v", err)
			return
		}

		var errUnknownEndpoint sdkerrors.ErrUnknownEndpoint
		ok = errors.As(err, &errUnknownEndpoint)
		if ok {
			logrus.Debugf("skipping request: %v", err)
			return
		}

		dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
		logrus.Errorf("Listing %s failed:\n%s", resourceType, dump)
		return
	}

	for _, r := range rs {
		state := queue.ItemStateNew
		reg := resource.GetRegistration(resourceType)
		if len(reg.DependsOn) > 0 {
			state = queue.ItemStateNewDependency
		}

		i := &queue.Item{
			Resource: r,
			State:    state,
			Type:     resourceType,
			Owner:    owner,
			Opts:     opts,
		}
		s.Items <- i
	}
}
