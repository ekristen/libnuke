package resource

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stevenle/topsort"
)

// Scope is a string in which resources are grouped against, this is meant for upstream tools to define their
// own scopes if the DefaultScope is not sufficient. For example Azure has multiple levels of scoping for resources,
// whereas AWS does not.
type Scope string

// DefaultScope is the default scope which resources are registered against if no other scope is provided
const DefaultScope Scope = "default"

// Registrations is a map of resource type to registration
type Registrations map[string]Registration

// Listers is a map of resource type to lister
type Listers map[string]Lister

// resourceListers is a global variable of all registered resource listers
var resourceListers = make(Listers)

// registrations is a global variable of all registrations for resources
var registrations = make(Registrations)

// graph is a global variable of the graph of resource dependencies
var graph = topsort.NewGraph()

// Registration is a struct that contains the information needed to register a resource lister
type Registration struct {
	Name   string
	Scope  Scope
	Lister Lister

	// DependsOn is a list of resource types that this resource type depends on. This is used to determine
	// the order in which resources are deleted. For example, a VPC depends on subnets, so we want to delete
	// the subnets before we delete the VPC. This is a Resource level dependency, not a resource instance (i.e. all
	// subnets must be deleted before any VPC can be deleted, not just the subnets that are associated with the VPC).
	DependsOn []string

	// DeprecatedAliases is a list of deprecated aliases for the resource type, usually misspellings or old names
	// that have been replaced with a new resource type. This is used to map the old resource type to the new
	// resource type. This is used in the config package to resolve any deprecated resource types and provide
	// notifications to the user.
	DeprecatedAliases []string

	// AlternativeResource is used to determine if there's an alternative resource type to use. The primary use case
	// for this is AWS Cloud Control API, where we want to use the Cloud Control API resource type instead of the
	// default resource type. However, any resource that uses a different API to manage the same resource can use this
	// field.
	AlternativeResource string
}

// Lister is an interface that represents a resource that can be listed
type Lister interface {
	List(ctx context.Context, opts interface{}) ([]Resource, error)
}

// RegisterOption is a function that can be used to manipulate the lister for a given resource type at
// registration time
type RegisterOption func(name string, lister Lister)

// Register registers a resource lister with the registry
func Register(r Registration, opts ...RegisterOption) {
	if r.Scope == "" {
		r.Scope = DefaultScope
	}

	if _, exists := registrations[r.Name]; exists {
		panic(fmt.Sprintf("a resource with the name %s already exists", r.Name))
	}

	logrus.WithField("name", r.Name).Trace("registered resource lister")

	registrations[r.Name] = r
	resourceListers[r.Name] = r.Lister

	graph.AddNode(r.Name)
	if len(r.DependsOn) == 0 {
		// Note: AddEdge will never through an error
		_ = graph.AddEdge("root", r.Name)
	}
	for _, dep := range r.DependsOn {
		// Note: AddEdge will never through an error
		_ = graph.AddEdge(dep, r.Name)
	}

	for _, opt := range opts {
		opt(r.Name, r.Lister)
	}
}

// ClearRegistry clears the registry of all registrations
// Designed for use for unit tests, not for production code. Only use if you know what you are doing.
func ClearRegistry() {
	registrations = make(Registrations)
	resourceListers = make(Listers)
	graph = topsort.NewGraph()
}

func GetListers() (listers Listers) {
	listers = make(Listers)
	for name, r := range registrations {
		listers[name] = r.Lister
	}
	return listers
}

// GetRegistration returns the registration for the given resource type
func GetRegistration(name string) Registration {
	return registrations[name]
}

// GetListersV2 returns a map of listers based on graph top sort order
func GetListersV2() (listers Listers) {
	listers = make(Listers)
	sorted, err := graph.TopSort("root")
	if err != nil {
		panic(err)
	}
	for _, name := range sorted {
		if name == "root" {
			continue
		}
		r := registrations[name]
		listers[name] = r.Lister
	}

	return listers
}

// GetListersForScope returns a map of listers for a particular scope that they've been grouped by
func GetListersForScope(scope Scope) (listers Listers) {
	listers = make(Listers)
	for name, r := range registrations {
		if r.Scope == scope {
			listers[name] = r.Lister
		}
	}
	return listers
}

// GetNames provides a string slice of all lister names that have been registered
func GetNames() []string {
	var names []string
	for resourceType := range GetListersV2() {
		names = append(names, resourceType)
	}

	return names
}

// GetNamesForScope provides a string slice of all listers for a particular scope
func GetNamesForScope(scope Scope) []string {
	var names []string
	for resourceType := range GetListersForScope(scope) {
		names = append(names, resourceType)
	}
	return names
}

// GetLister gets a specific lister by name
func GetLister(name string) Lister {
	return resourceListers[name]
}

func GetAlternativeResourceTypeMapping() map[string]string {
	mapping := make(map[string]string)
	for _, r := range registrations {
		if r.AlternativeResource != "" {
			mapping[r.Name] = r.AlternativeResource
		}
	}
	return mapping
}

// GetDeprecatedResourceTypeMapping returns a map of deprecated resource types to their replacement
func GetDeprecatedResourceTypeMapping() map[string]string {
	mapping := make(map[string]string)
	for _, r := range registrations {
		for _, alias := range r.DeprecatedAliases {
			mapping[alias] = r.Name
		}
	}
	return mapping
}
