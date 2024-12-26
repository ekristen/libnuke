// Package registry provides a way to register resources and their listers and obtain them after the fact. The registry
// is currently deeply embedded with the other packages and how they access specific aspects of a resource.
package registry

import (
	"context"
	"fmt"
	"sort"

	"github.com/mb0/glob"
	"github.com/sirupsen/logrus"
	"github.com/stevenle/topsort"

	"github.com/ekristen/libnuke/pkg/resource"
)

// Scope is a string in which resources are grouped against, this is meant for upstream tools to define their
// own scopes if the DefaultScope is not sufficient. For example Azure has multiple levels of scoping for resources,
// whereas AWS does not.
type Scope string

// DefaultScope is the default scope which resources are registered against if no other scope is provided
const DefaultScope Scope = "default"

// Registrations is a map of resource type to registration
type Registrations map[string]*Registration

// Listers is a map of resource type to lister
type Listers map[string]Lister

// resourceListers is a global variable of all registered resource listers
var resourceListers = make(Listers)

// registrations is a global variable of all registrations for resources
var registrations = make(Registrations)

// alternatives is a global variable of all alternative resource types
var alternatives = make(map[string]string)

// graph is a global variable of the graph of resource dependencies
var graph = topsort.NewGraph()

// Registration is a struct that contains the information needed to register a resource lister
type Registration struct {
	// Name is the name of the resource type
	Name string

	// Scope is the scope of the resource type, if left empty it'll default to DefaultScope. It's simple a string
	// designed to group resource types together. The primary use case is for Azure, it needs resources scoped to
	// different levels, whereas AWS has simply Account level.
	Scope Scope

	// Resource is the resource type that the lister is going to list. This is a struct that implements the Resource
	// interface. This is primarily used to generate documentation by parsing the structs properties and generating
	// markdown documentation.
	// Note: it is a interface{} because we are going to inspect it, we do not need to actually call any methods on it.
	Resource interface{}

	// Lister is the lister for the resource type, it is a struct with a method called List that returns a slice
	// of resources. The lister is responsible for filtering out any resources that should not be deleted because they
	// are ineligible for deletion. For example, built in resources that cannot be deleted.
	Lister Lister

	// Settings allows for resources to define settings that can be configured by the calling tool to change the
	// behavior of the resource. For example, EC2 and RDS Instances have Deletion Protection, this allows the resource
	// to define a setting that can be configured by the calling tool to enable/disable deletion protection.
	Settings []string

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
	List(ctx context.Context, opts interface{}) ([]resource.Resource, error)
}

// ListerWithClose is an interface that represents a lister that can be closed. Use Case: GCP clients need to be closed.
type ListerWithClose interface {
	Close()
}

// RegisterOption is a function that can be used to manipulate the lister for a given resource type at
// registration time
type RegisterOption func(name string, lister Lister)

// Register registers a resource lister with the registry
func Register(r *Registration) {
	if r.Scope == "" {
		r.Scope = DefaultScope
	}

	if _, exists := registrations[r.Name]; exists {
		panic(fmt.Sprintf("a resource with the name %s already exists", r.Name))
	}

	if r.AlternativeResource != "" {
		if _, exists := alternatives[r.AlternativeResource]; exists {
			panic(fmt.Sprintf("an alternative resource mapping for %s already exists", r.AlternativeResource))
		}

		alternatives[r.AlternativeResource] = r.Name
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
}

// GetRegistrations returns all registrations
func GetRegistrations() Registrations {
	return registrations
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
func GetRegistration(name string) *Registration {
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

// ExpandNames takes a list of names and expands them based on a wildcard and returns all the names that match
func ExpandNames(names []string) []string {
	var expandedNames []string
	registeredNames := GetNames()

	for _, name := range names {
		matches, _ := glob.GlobStrings(registeredNames, name)
		if matches == nil {
			logrus.
				WithField("handler", "ExpandNames").
				WithField("name", name).
				Trace("no expansion for name")

			expandedNames = append(expandedNames, name)
			continue
		}

		logrus.
			WithField("handler", "ExpandNames").
			WithField("name", name).
			WithField("matches", matches).
			Trace("expanded name")

		expandedNames = append(expandedNames, matches...)
	}

	// Ensure predictable order
	sort.Strings(expandedNames)

	return expandedNames
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

// GetAlternativeResourceTypeMapping returns a map of resource types to their alternative resource type. The primary
// use case is used to map resource types to their alternative AWS Cloud Control resource type. This allows each
// resource to define what it's alternative resource type is instead of trying to track them in a single place.
func GetAlternativeResourceTypeMapping() map[string]string {
	mapping := make(map[string]string)
	for _, r := range registrations {
		if r.AlternativeResource != "" {
			mapping[r.Name] = r.AlternativeResource
		}
	}
	return mapping
}

// GetDeprecatedResourceTypeMapping returns a map of deprecated resource types to their replacement. The primary use
// case is used to map deprecated resource types to their replacement in the config package. This allows us to
// provide notifications to the user that they are using a deprecated resource type and should update their config.
// This allow allows each resource to define it's DeprecatedAliases instead of trying to track them in a single place.
func GetDeprecatedResourceTypeMapping() map[string]string {
	mapping := make(map[string]string)
	for _, r := range registrations {
		for _, alias := range r.DeprecatedAliases {
			mapping[alias] = r.Name
		}
	}
	return mapping
}
