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
	Items          chan *queue.Item    `hash:"ignore"`
	Semaphore      *semaphore.Weighted `hash:"ignore"`
	ResourceTypes  []string
	Options        interface{}
	Owner          string
	MutateOptsFunc MutateOptsFunc `hash:"ignore"`
}

// MutateOptsFunc is a function that can mutate the Options for a given resource type. This is useful for when you
// need to pass in a different set of Options for a given resource type. For example, AWS nuke needs to be able to
// populate the region and session for a given resource type give that it might only exist in us-east-1.
type MutateOptsFunc func(opts interface{}, resourceType string) interface{}

// NewScanner creates a new scanner for the given resource types.
func NewScanner(owner string, resourceTypes []string, opts interface{}) *Scanner {
	return &Scanner{
		Items:         make(chan *queue.Item, 10000),
		Semaphore:     semaphore.NewWeighted(ScannerParallelQueries),
		ResourceTypes: resourceTypes,
		Options:       opts,
		Owner:         owner,
	}
}

type IScanner interface {
	Run(resourceTypes []string)
	list(resourceType string)
}

// RegisterMutateOptsFunc registers a mutate Options function for the scanner. The mutate Options function is called
// for each resource type that is being scanned. This allows you to mutate the Options for a given resource type.
func (s *Scanner) RegisterMutateOptsFunc(morph MutateOptsFunc) error {
	if s.MutateOptsFunc != nil {
		return fmt.Errorf("MutateOptsFunc already registered")
	}
	s.MutateOptsFunc = morph
	return nil
}

// Run starts the scanner and runs the lister for each resource type.
func (s *Scanner) Run(ctx context.Context) error {
	for _, resourceType := range s.ResourceTypes {
		if err := s.Semaphore.Acquire(ctx, 1); err != nil {
			return err
		}

		opts := s.Options
		if s.MutateOptsFunc != nil {
			opts = s.MutateOptsFunc(opts, resourceType)
		}

		go s.list(ctx, s.Owner, resourceType, opts)
	}

	// Wait for all routines to finish.
	if err := s.Semaphore.Acquire(ctx, ScannerParallelQueries); err != nil {
		return err
	}

	close(s.Items)

	return nil
}

func (s *Scanner) list(ctx context.Context, owner, resourceType string, opts interface{}) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v\n\n%s", r.(error), string(debug.Stack()))
			dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
			logrus.Errorf("Listing %s failed:\n%s", resourceType, dump)
		}
	}()

	defer s.Semaphore.Release(1)

	lister := resource.GetLister(resourceType)
	var rs []resource.Resource

	rs, err := lister.List(ctx, opts)
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
		logrus.WithError(err).Errorf("Listing %s failed:\n%s", resourceType, dump)
		return
	}

	for _, r := range rs {
		i := &queue.Item{
			Resource: r,
			State:    queue.ItemStateNew,
			Type:     resourceType,
			Owner:    owner,
			Opts:     opts,
		}
		s.Items <- i
	}
}
