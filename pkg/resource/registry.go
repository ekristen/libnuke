package resource

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stevenle/topsort"
)

// Scope is a string in which resources are grouped against
type Scope string

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
	Name      string
	Scope     Scope
	Lister    Lister
	DependsOn []string
}

// Lister is an interface that represents a resource that can be listed
type Lister interface {
	List(opts interface{}) ([]Resource, error)
}

// RegisterOption is a function that can be used to manipulate the lister for a given resource type at
// registration time
type RegisterOption func(name string, lister Lister)

// Register registers a resource lister with the registry
func Register(r Registration, opts ...RegisterOption) {
	if r.Scope == "" {
		panic(fmt.Errorf("scope must be set"))
	}

	_, exists := registrations[r.Name]
	if exists {
		panic(fmt.Sprintf("a resource with the name %s already exists", r.Name))
	}

	logrus.WithField("name", r.Name).Trace("registered resource lister")

	registrations[r.Name] = r
	resourceListers[r.Name] = r.Lister

	graph.AddNode(r.Name)
	if len(r.DependsOn) == 0 {
		err := graph.AddEdge("root", r.Name)
		if err != nil {
			panic(err)
		}
	}
	for _, dep := range r.DependsOn {
		err := graph.AddEdge(dep, r.Name)
		if err != nil {
			panic(err)
		}
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
