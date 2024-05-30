package soop

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewSupervisorNode tests the creation of a new SupervisorNode.
func TestSupervisor_NewSupervisorNode(t *testing.T) {
	ctx := context.Background()
	name := "TestSupervisor"
	buffSize := 10
	handler := func(err error) error { return err }

	supervisor := NewSupervisor[int, int](ctx, name, buffSize, handler)

	assert.NotNil(t, supervisor)
	assert.Contains(t, supervisor.GetName(), name, "Expected node name to contain the assigned name")
	assert.Equal(t, NodeTypeSupervisor, supervisor.nodeType)
	assert.NotNil(t, supervisor.children)
}

// TestSetEventOutChan tests setting the event output channel.
func TestSupervisor_SetEventOutChan(t *testing.T) {
	supervisor := NewSupervisor[int, int](context.Background(), "TestSupervisor", 10, nil)
	eventOutChan := make(chan error)
	supervisor.SetEventOutChan(eventOutChan)
	assert.Equal(t, eventOutChan, supervisor.eventOutChan)
}

// TestAddChild tests adding a child supervisor node to the supervisor.
func TestSupervisor_AddChildSupervisor(t *testing.T) {
	supervisor := NewSupervisor[int, int](context.Background(), "TestSupervisor", 10, nil)
	child := NewSupervisor[int, int](supervisor.ctx, "ChildSupervisorNode", 10, nil)

	err := supervisor.AddChild(child)
	assert.NoError(t, err)
	assert.Contains(t, supervisor.GetChildren(), child.GetID())
}

// TestAddChild tests adding a child pool supervisor node to the supervisor.
func TestSupervisor_AddChildPoolSupervisor(t *testing.T) {
	supervisor := NewSupervisor[int, int](context.Background(), "TestSupervisor", 10, nil)
	child := NewPoolSupervisor[int, int](supervisor.ctx, "ChildPoolSupervisorNode", 10, nil)

	err := supervisor.AddChild(child)
	assert.NoError(t, err)
	assert.Contains(t, supervisor.GetChildren(), child.GetID())
}

// TestGetID tests getting the ID of the supervisor node.
func TestSupervisor_GetID(t *testing.T) {
	supervisor := NewSupervisor[int, int](context.Background(), "TestSupervisor", 10, nil)
	assert.Equal(t, supervisor.ID, supervisor.GetID())
}

// TestGetName tests getting the name of the supervisor node.
func TestSupervisor_GetName(t *testing.T) {
	supervisor := NewSupervisor[int, int](context.Background(), "TestSupervisor", 10, nil)
	assert.Contains(t, supervisor.GetName(), "supervisor_TestSupervisor_", "Expected node name to contain the node type and assigned name")
}

// TestGetChildren tests getting the child nodes of the supervisor.
func TestSupervisor_GetChildren(t *testing.T) {
	supervisor := NewSupervisor[int, int](context.Background(), "TestSupervisor", 10, nil)
	for i := 0; i < 5; i++ {
		child := NewSupervisor[int, int](supervisor.ctx, fmt.Sprintf("ChildSupervisorNode-%d", i), 10, nil)
		err := supervisor.AddChild(child)
		assert.NoError(t, err)
	}

	assert.Equal(t, 5, len(supervisor.GetChildren()))
}

// TestHandle tests handling an event in the supervisor node.
func TestSupervisor_Handle(t *testing.T) {
	handler := func(err error) error { return err }
	supervisor := NewSupervisor[int, int](context.Background(), "TestSupervisor", 10, handler)
	eventOutChan := make(chan error, 1)
	supervisor.SetEventOutChan(eventOutChan)

	testError := errors.New("test error")
	supervisor.Handle(testError)

	assert.Equal(t, testError, <-eventOutChan)
}

// TestStart tests starting the supervisor and its children.
func TestSupervisor_Start(t *testing.T) {
	supervisor := NewSupervisor[int, int](context.Background(), "TestSupervisor", 10, nil)
	for i := 0; i < 5; i++ {
		child := NewSupervisor[int, int](supervisor.ctx, fmt.Sprintf("ChildSupervisorNode-%d", i), 10, nil)
		err := supervisor.AddChild(child)
		assert.NoError(t, err)
	}

	err := supervisor.Start()
	assert.NoError(t, err)
}
