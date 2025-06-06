// Package scanner provides a mechanism for scanning resources and adding them to the item queue for processing. The
// scope of the scanner is determined by the resource types that are passed to it. The scanner will then run the lister
// for each resource type and add the resources to the item queue for processing.
package scanner

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	liberrors "github.com/ekristen/libnuke/pkg/errors"

	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/utils"
)

// DefaultParallelQueries is the number of parallel queries to run at any given time for a scanner.
const DefaultParallelQueries = 16

// DefaultQueueSize is the default size of the item queue for a scanner.
const DefaultQueueSize = 50000

// Scanner is collection of resource types that will be scanned for existing resources and added to the
// item queue for processing. These items will be filtered and then processed.
type Scanner struct {
	Items           chan *queue.Item    `hash:"ignore"`
	semaphore       *semaphore.Weighted `hash:"ignore"`
	ResourceTypes   []string
	Options         interface{}
	Owner           string
	mutateOptsFunc  MutateOptsFunc `hash:"ignore"`
	parallelQueries int64
	logger          *logrus.Logger
}

// MutateOptsFunc is a function that can mutate the Options for a given resource type. This is useful for when you
// need to pass in a different set of Options for a given resource type. For example, AWS nuke needs to be able to
// populate the region and session for a given resource type give that it might only exist in us-east-1.
type MutateOptsFunc func(opts interface{}, resourceType string) interface{}

// Config is the configuration for a scanner.
type Config struct {
	Owner           string
	ResourceTypes   []string
	Opts            interface{}
	QueueSize       int
	ParallelQueries int64
	Logger          *logrus.Logger
}

// New creates a new scanner for the given resource types.
func New(cfg *Config) (*Scanner, error) {
	if cfg.Owner == "" {
		return nil, fmt.Errorf("owner must be set")
	}
	if cfg.QueueSize == 0 {
		cfg.QueueSize = DefaultQueueSize
	}
	if cfg.ParallelQueries == 0 {
		cfg.ParallelQueries = DefaultParallelQueries
	}
	if cfg.Logger == nil {
		cfg.Logger = logrus.StandardLogger()
	}

	return &Scanner{
		Items:           make(chan *queue.Item, cfg.QueueSize),
		semaphore:       semaphore.NewWeighted(cfg.ParallelQueries),
		ResourceTypes:   cfg.ResourceTypes,
		Options:         cfg.Opts,
		Owner:           cfg.Owner,
		parallelQueries: cfg.ParallelQueries,
		logger:          cfg.Logger,
	}, nil
}

type IScanner interface {
	Run(resourceTypes []string)
	list(resourceType string)
}

// RegisterMutateOptsFunc registers a mutate Options function for the scanner. The mutate Options function is called
// for each resource type that is being scanned. This allows you to mutate the Options for a given resource type.
func (s *Scanner) RegisterMutateOptsFunc(morph MutateOptsFunc) error {
	if s.mutateOptsFunc != nil {
		return fmt.Errorf("mutateOptsFunc already registered")
	}
	s.mutateOptsFunc = morph
	return nil
}

// SetParallelQueries changes the number of parallel queries to run at any given time from the default for the scanner.
func (s *Scanner) SetParallelQueries(parallelQueries int64) {
	s.parallelQueries = parallelQueries
	s.semaphore = semaphore.NewWeighted(s.parallelQueries)
}

// SetLogger sets the logger for the scanner.
func (s *Scanner) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

// Run starts the scanner and runs the lister for each resource type.
func (s *Scanner) Run(ctx context.Context) error {
	for _, resourceType := range s.ResourceTypes {
		if err := s.semaphore.Acquire(ctx, 1); err != nil {
			return err
		}

		opts := s.Options
		if s.mutateOptsFunc != nil {
			opts = s.mutateOptsFunc(opts, resourceType)
		}

		go s.list(ctx, s.Owner, resourceType, opts)
	}

	// Wait for all routines to finish.
	if err := s.semaphore.Acquire(ctx, s.parallelQueries); err != nil {
		return err
	}

	close(s.Items)

	return nil
}

func (s *Scanner) list(ctx context.Context, owner, resourceType string, opts interface{}) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := logrus.WithField("resource_type", resourceType).WithField("owner", owner)

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v\n\n%s", r.(error), string(debug.Stack()))
			dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
			logger.Errorf("listing failed:\n%s", dump)
		}
	}()

	defer s.semaphore.Release(1)

	lister := registry.GetLister(resourceType)
	var rs []resource.Resource

	if lister == nil {
		logger.Error("lister for resource type not found")
		return
	}

	logger.Debug("attempting to run lister")

	rs, err := lister.List(ctx, opts)
	if err != nil {
		var errSkipRequest liberrors.ErrSkipRequest
		ok := errors.As(err, &errSkipRequest)
		if ok {
			logger.Debugf("skipping request: %v", err)
			return
		}

		var errUnknownEndpoint liberrors.ErrUnknownEndpoint
		ok = errors.As(err, &errUnknownEndpoint)
		if ok {
			logger.Debugf("skipping request: %v", err)
			return
		}

		dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
		logger.WithError(err).Errorf("listing failed:\n%s", dump)
		return
	}

	logger.WithField("count", len(rs)).Debugf("listing complete")

	queueFullWarned := false
	for _, r := range rs {
		i := &queue.Item{
			Resource: r,
			State:    queue.ItemStateNew,
			Type:     resourceType,
			Owner:    owner,
			Opts:     opts,
			Logger:   s.logger,
		}

		itemHook, ok := r.(resource.QueueItemHook)
		if ok {
			itemHook.BeforeEnqueue(i)
		}

		select {
		case s.Items <- i:
			// successfully enqueued
		default:
			if !queueFullWarned {
				logger.Warn("item queue is full, not all resources will be enqueued")
				queueFullWarned = true
			}
			break
		}
		if queueFullWarned {
			break
		}
	}

	logger.Debugf("resources enqueue complete")
}
