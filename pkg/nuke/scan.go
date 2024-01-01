package nuke

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	sdkerrors "github.com/ekristen/cloud-nuke-sdk/pkg/errors"
	"github.com/ekristen/cloud-nuke-sdk/pkg/queue"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/utils"
)

const ScannerParallelQueries = 16

type Scanner struct {
	Items     chan *queue.Item
	semaphore *semaphore.Weighted

	resourceTypes []string
	options       interface{}
	owner         string
}

func NewScanner(owner string, resourceTypes []string, opts interface{}) *Scanner {
	return &Scanner{
		Items:         make(chan *queue.Item, 0),
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

func (s *Scanner) Run() {
	ctx := context.Background()

	for _, resourceType := range s.resourceTypes {
		s.semaphore.Acquire(ctx, 1)
		go s.list(s.owner, resourceType, s.options)
	}

	// Wait for all routines to finish.
	s.semaphore.Acquire(ctx, ScannerParallelQueries)

	close(s.Items)
}

func (s *Scanner) list(owner, resourceType string, opts interface{}) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v\n\n%s", r.(error), string(debug.Stack()))
			dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
			log.Errorf("Listing %s failed:\n%s", resourceType, dump)
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
			log.Debugf("skipping request: %v", err)
			return
		}

		var errUnknownEndpoint sdkerrors.ErrUnknownEndpoint
		ok = errors.As(err, &errUnknownEndpoint)
		if ok {
			log.Warnf("skipping request: %v", err)
			return
		}

		dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
		log.Errorf("Listing %s failed:\n%s", resourceType, dump)
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
