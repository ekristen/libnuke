package resource

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stevenle/topsort"
)

type Scope string

type Registrations map[string]Registration
type Listers map[string]Lister

var resourceListers = make(Listers)
var registrations = make(Registrations)
var graph = topsort.NewGraph()

type Registration struct {
	Name      string
	Scope     Scope
	Lister    Lister
	DependsOn []string
}

type Lister interface {
	List(opts interface{}) ([]Resource, error)
}

func Register(r Registration) {
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
}

func GetListers() (listers Listers) {
	listers = make(Listers)
	for name, r := range registrations {
		listers[name] = r.Lister
	}
	return listers
}

func GetRegistration(name string) Registration {
	return registrations[name]
}

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

func GetListersForScope(scope Scope) (listers Listers) {
	listers = make(Listers)
	for name, r := range registrations {
		if r.Scope == scope {
			listers[name] = r.Lister
		}
	}
	return listers
}

func GetNames() []string {
	var names []string
	for resourceType := range GetListersV2() {
		names = append(names, resourceType)
	}

	return names
}

func GetNamesForScope(scope Scope) []string {
	var names []string
	for resourceType := range GetListersForScope(scope) {
		names = append(names, resourceType)
	}
	return names
}

func GetLister(name string) Lister {
	return resourceListers[name]
}
