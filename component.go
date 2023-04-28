package framework

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

var (
	alreadyRegisteredError = errors.New("component a already has a dependency on component b")
	dependencyCycleError   = errors.New("Dependency cycle. component b already depends on component a.")
)

// Component defines the functionality that is required to start and stop a component when a service is started
type Component interface {
	// Will be called to start the component
	// Once start returns without error, the program will continue with the next component.
	// An error response stops the initialization process
	// Context can be provided to handle timeouts
	Start(ctx context.Context) error
	// Will be called to request the component to stop
	// It must block until the component is completely stopped
	// An error response will be logged, but the program will continue to stop other dependencies
	// Context can be provided to handle timeouts
	Stop(ctx context.Context) error
}

// Service contains all of the components that you wish to have running
type Service struct {
	components          []Component
	dependantComponents map[Component][]Component
	startedComponents   map[Component]bool

	closingStack *Stack
}

// NewService creates a new service
func NewService(components []Component) *Service {

	return &Service{
		components:          components,
		dependantComponents: make(map[Component][]Component),
		startedComponents:   make(map[Component]bool),
		closingStack:        NewStack(),
	}
}

// RegisterDependentComponents will register that component a depends on component b
func (s *Service) RegisterDependentComponents(a, b Component) error {
	// check that b doesn't depend on a to avoid cyclic dependencies
	dependencies, ok := s.dependantComponents[b]
	if ok {
		for _, dep := range dependencies {
			if dep == a {
				return dependencyCycleError
			}
		}
	}

	// check to see if the dependency has already been set up
	existingDependencies, ok := s.dependantComponents[a]
	if ok {
		for _, dependency := range existingDependencies {
			if dependency == b {
				return alreadyRegisteredError
			}
		}
	}

	s.dependantComponents[a] = append(s.dependantComponents[a], b)

	return nil
}

// Start will start the components that have been added. It will ensure that components that are dependencies of
// other components are started before the components that depend on them. If any component fails to start, an error
// will be returned without continuing to start the rest of the components.
func (s *Service) Start(ctx context.Context) error {
	for _, component := range s.components {
		err := s.startComponent(ctx, component)
		if err != nil {
			return errors.Wrap(err, "failed to start component")
		}
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) {
	s.closingStack.Close(ctx)
}

func (s *Service) startComponent(ctx context.Context, component Component) error {
	// check the component hasn't already been started due to it being a dependent component
	if _, ok := s.startedComponents[component]; ok {
		return nil
	}

	// first check for components that this component depends upon so that they can be started first
	depComponents := s.checkForDependantComponents(component)

	if len(depComponents) > 0 {
		for _, dependency := range depComponents {
			err := s.startComponent(ctx, dependency)
			if err != nil {
				return errors.Wrap(err, "failed to start component")
			}
		}
	}

	err := component.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start component")
	}

	// We need to stop components in the opposite order that they are started. This is to ensure that
	// dependency components are shut down after components that depend on them. By adding them to the
	// closing stack here (which closes them in reverse order of being added), we ensure that dependency
	// components are added before the components that depend on them, thus being shut down after the
	// components that depend on them
	s.closingStack.Add("something", CloseFunc(func(ctx context.Context) {
		err := component.Stop(ctx)
		if err != nil {
			fmt.Printf("failed to stop component: %s\n", err)
		}
	}))

	s.startedComponents[component] = true
	return nil
}

func (s *Service) checkForDependantComponents(componentToCheck Component) []Component {
	components := make([]Component, 0, 0)
	dependencies, ok := s.dependantComponents[componentToCheck]
	if !ok {
		// component doesn't depend on anything
		return nil
	}

	for _, dependency := range dependencies {
		// only add the dependency to the dependencies to return if it hasn't already been started
		if _, ok := s.startedComponents[dependency]; !ok {
			components = append(components, dependency)
		}
	}

	return components
}
