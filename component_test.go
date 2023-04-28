package framework

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	a := &CompA{}
	b := &CompB{}

	service := NewService([]Component{a, b})

	err := service.Start(context.Background())
	require.NoError(t, err)

	// check that both components were started and that start() was only called once
	assert.True(t, a.started)
	assert.Equal(t, 1, a.startCallCount)
	assert.True(t, b.started)
	assert.Equal(t, 1, b.startCallCount)

	// check that component a was started before b since there's no dependencies and a was the first in
	// the slice of components the service received

	// NOTE: Technically this shouldn't matter since the caller isn't specifying the order in which the components should be
	// started, just the components to start. However for the sake of this POC, it proves that we aren't calling a dependent
	// component first
	require.True(t, a.startedAt.Before(b.startedAt))
}

func TestNewServiceWithDependents(t *testing.T) {
	a := &CompA{}
	b := &CompB{}

	// create a service with the components
	service := NewService([]Component{a, b})

	// configure a to depend on b
	err := service.RegisterDependentComponents(a, b)
	require.NoError(t, err)

	err = service.Start(context.Background())
	require.NoError(t, err)

	// check that both components were started and that start() was only called once
	assert.True(t, a.started)
	assert.Equal(t, 1, a.startCallCount)
	assert.True(t, b.started)
	assert.Equal(t, 1, b.startCallCount)

	// check that component b was started before a since a depends on b
	require.True(t, b.startedAt.Before(a.startedAt))
}

func TestNewServiceWithCyclicDependency(t *testing.T) {
	a := &CompA{}
	b := &CompB{}

	service := NewService([]Component{a, b})

	// first configure a to depend on b
	err := service.RegisterDependentComponents(a, b)
	require.NoError(t, err)

	// now try to configure b to depend on a
	err = service.RegisterDependentComponents(b, a)
	require.Error(t, err)
	assert.Equal(t, dependencyCycleError, err)
}

func TestNewServiceWithExistingDependency(t *testing.T) {
	a := &CompA{}
	b := &CompB{}

	service := NewService([]Component{a, b})

	// configure a to depend on b
	err := service.RegisterDependentComponents(a, b)
	require.NoError(t, err)

	// try to configure a to depend on b again
	err = service.RegisterDependentComponents(a, b)
	require.Error(t, err)
	assert.Equal(t, alreadyRegisteredError, err)
}

func TestNewServiceWithMultiLayerDependencies(t *testing.T) {
	a := &CompA{}
	b := &CompB{}
	c := &CompC{}

	// create a service with the components
	service := NewService([]Component{a, b, c})

	// configure a to depend on b
	err := service.RegisterDependentComponents(a, b)
	require.NoError(t, err)

	// configure b to depend on c
	err = service.RegisterDependentComponents(b, c)
	require.NoError(t, err)

	err = service.Start(context.Background())
	require.NoError(t, err)

	// check that all components were started and that start() was only called once
	assert.True(t, a.started)
	assert.Equal(t, 1, a.startCallCount)
	assert.True(t, b.started)
	assert.Equal(t, 1, b.startCallCount)
	assert.True(t, c.started)
	assert.Equal(t, 1, c.startCallCount)

	// check that component b was started before a since a depends on b
	require.True(t, b.startedAt.Before(a.startedAt))

	// check that component c was started before b since b depends on c
	require.True(t, c.startedAt.Before(b.startedAt))
}

func TestServiceStop(t *testing.T) {
	a := &CompA{}
	b := &CompB{}

	// create a service with the components
	service := NewService([]Component{a, b})

	// configure a to depend on b
	err := service.RegisterDependentComponents(a, b)
	require.NoError(t, err)

	err = service.Start(context.Background())
	require.NoError(t, err)

	service.Stop(context.Background())

	// check that both components were stopped and that stop() was only called once
	assert.False(t, a.started)
	assert.Equal(t, 1, a.stopCallCount)
	assert.False(t, b.started)
	assert.Equal(t, 1, b.stopCallCount)

	// check that component b was stopped after a since a depends on b and so a should be shut down first
	require.True(t, b.stoppedAt.After(a.stoppedAt))
}

func TestServiceStopMultiLayeredDependencies(t *testing.T) {
	a := &CompA{}
	b := &CompB{}
	c := &CompC{}

	// create a service with the components
	service := NewService([]Component{a, b, c})

	// configure a to depend on b
	err := service.RegisterDependentComponents(a, b)
	require.NoError(t, err)

	// configure component a to also depend on c
	err = service.RegisterDependentComponents(a, c)
	require.NoError(t, err)

	// configure component b to depend on component c
	err = service.RegisterDependentComponents(b, c)
	require.NoError(t, err)

	err = service.Start(context.Background())
	require.NoError(t, err)

	service.Stop(context.Background())

	// check that all components were stopped and that stop() was only called once
	assert.False(t, a.started)
	assert.Equal(t, 1, a.stopCallCount)
	assert.False(t, b.started)
	assert.Equal(t, 1, b.stopCallCount)
	assert.False(t, c.started)
	assert.Equal(t, 1, c.stopCallCount)

	// check that component b was stopped after a since a depends on b and so a should be shut down first
	require.True(t, b.stoppedAt.After(a.stoppedAt))

	// check that component c was stopped after b since b depends on c and so b should be shut down first
	require.True(t, c.stoppedAt.After(b.stoppedAt))

	// check that component c was stopped after a since a depends on c and so a should be shut down first
	require.True(t, c.stoppedAt.After(a.stoppedAt))
}

// Some dummy components

type CompA struct {
	startCallCount int
	stopCallCount  int
	started        bool
	startedAt      time.Time
	stoppedAt      time.Time
}

func (a *CompA) Start(ctx context.Context) error {
	a.started = true
	a.startCallCount++

	// simulate some set up stuff like connecting to NATs or starting an HTTP server
	time.Sleep(time.Millisecond * 100)
	a.startedAt = time.Now()

	return nil
}

func (a *CompA) Stop(ctx context.Context) error {
	a.started = false
	a.stopCallCount++

	// simulate some stuff like closing connections
	time.Sleep(time.Millisecond * 100)
	a.stoppedAt = time.Now()
	return nil
}

type CompB struct {
	startCallCount int
	stopCallCount  int
	started        bool
	startedAt      time.Time
	stoppedAt      time.Time
}

func (b *CompB) Start(ctx context.Context) error {
	b.startCallCount++
	b.started = true

	// simulate some set up stuff like connecting to NATs or starting an HTTP server
	time.Sleep(time.Millisecond * 100)
	b.startedAt = time.Now()

	return nil
}

func (b *CompB) Stop(ctx context.Context) error {
	b.started = false
	b.stopCallCount++

	// simulate some stuff like closing connections
	time.Sleep(time.Millisecond * 100)
	b.stoppedAt = time.Now()
	return nil
}

type CompC struct {
	startCallCount int
	stopCallCount  int
	started        bool
	startedAt      time.Time
	stoppedAt      time.Time
}

func (c *CompC) Start(ctx context.Context) error {
	c.startCallCount++
	c.started = true

	// simulate some set up stuff like connecting to NATs or starting an HTTP server
	time.Sleep(time.Millisecond * 100)
	c.startedAt = time.Now()

	return nil
}

func (c *CompC) Stop(ctx context.Context) error {
	c.started = false
	c.stopCallCount++

	// simulate some stuff like closing connections
	time.Sleep(time.Millisecond * 100)
	c.stoppedAt = time.Now()
	return nil
}
