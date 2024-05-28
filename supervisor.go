package soop

import (
	"context"
	"errors"
	"fmt"
)

type Supervisor[I, O any] interface {
	GetID() uint64
	GetName() string
	Start() error
	Stop() error
	GetChildren() map[uint64]Node
	SetEventOutChan(chan error)
}

// SupervisorNode represents a supervisor node that manages child nodes.
type SupervisorNode[I, O any] struct {
	node
	children     map[uint64]Node
	eventInChan  chan error
	eventOutChan chan error
	handler      EventHandler
	ctx          context.Context
	cancelFunc   context.CancelFunc
}

// NewSupervisor creates a new supervisor node with children.
func NewSupervisor[I, O any](ctx context.Context, name string, buffSize int, handler EventHandler) *SupervisorNode[I, O] {
	newCtx, cancel := context.WithCancel(ctx)
	return &SupervisorNode[I, O]{
		node:       newNode(ctx, name, NodeTypeSupervisor, nil), // TODO: Add factory function
		ctx:        newCtx,
		cancelFunc: cancel,
		children:   make(map[uint64]Node),
	}
}

// SetEventOutChan sets the output event channel.
func (s *SupervisorNode[I, O]) SetEventOutChan(eventOutChan chan error) {
	s.eventOutChan = eventOutChan
}

// AddChild adds a child node to the supervisor.
func (s *SupervisorNode[I, O]) AddChild(node Node) error {
	if s.nodeType != NodeTypeDefault && s.nodeType != NodeTypeSupervisor {
		return NewError[error](s.ID, ErrorLevelError, fmt.Sprintf("Cannot add a child to a %s node", nodeTypeNames[s.nodeType]), nil, nil)
	}

	switch child := node.(type) {
	case Supervisor[I, O]:
		child.SetEventOutChan(s.eventInChan)
		s.children[node.GetID()] = child
	case PoolSupervisor[I, O]:
		child.SetEventOutChan(s.eventInChan)
		s.children[node.GetID()] = child
	default:
		return NewError[error](s.ID, ErrorLevelError, "Child node is not a supervisor", nil, nil)
	}

	return nil
}

// GetID returns the ID of the supervisor node.
func (s *SupervisorNode[I, O]) GetID() uint64 {
	return s.ID
}

// GetName returns the name of the supervisor node.
func (s *SupervisorNode[I, O]) GetName() string {
	return s.name
}

// GetChildren returns the child nodes.
func (s *SupervisorNode[I, O]) GetChildren() map[uint64]Node {
	return s.children
}

// Handle processes an error event.
func (s *SupervisorNode[I, O]) Handle(event error) {
	var customErr *Error[I]
	if errors.As(event, &customErr) {
		if customErr.Level == ErrorLevelCritical {
			// Remove child
			// TODO: Remove additional references, like children of a child, or nodes in a pool in a child pool supervisor.
			delete(s.children, customErr.Node)
		}
	}

	if s.eventOutChan != nil {
		s.eventOutChan <- s.handler(event)
	}
}

// Start starts the supervisor and its children.
func (s *SupervisorNode[I, O]) Start() error {
	for _, child := range s.children {
		err := child.Start()
		if err != nil {
			return err
		}
	}

	return nil
}
