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

type Registration struct {
	Name      string
	Scope     Scope
	Lister    Lister
	DependsOn []string
}

type ListerOpts struct {
}

type Lister func(lister ListerOpts) ([]Resource, error)

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
}

func GetListers() (listers Listers) {
	listers = make(Listers)
	for name, r := range registrations {
		listers[name] = r.Lister
	}
	return listers
}

func GetListersTS() {
	graph := topsort.NewGraph()

	for name := range registrations {
		graph.AddNode(name)
	}
	for name, r := range registrations {
		//if r.Scope == ResourceGroup {
		//	graph.AddEdge("ResourceGroup", name)
		//}
		for _, dep := range r.DependsOn {
			graph.AddEdge(name, dep)
		}
	}
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
	for resourceType := range GetListers() {
		names = append(names, resourceType)
	}

	return names
}

func GetNameForScope(scope Scope) []string {
	var names []string
	for resourceType := range GetListersForScope(scope) {
		names = append(names, resourceType)
	}
	return names
}

func GetLister(name string) Lister {
	return resourceListers[name]
}
