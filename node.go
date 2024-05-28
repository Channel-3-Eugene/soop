package soop

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type Node interface {
	GetID() uint64
	GetName() string
	Start() error
	Stop() error
}

type NodeType int

const (
	NodeTypeDefault NodeType = iota
	NodeTypeWorker
	NodeTypePool
	NodeTypePoolSupervisor
	NodeTypeSupervisor
)

var nodeTypeNames = map[NodeType]string{
	NodeTypeDefault:        "node",
	NodeTypeWorker:         "worker",
	NodeTypePool:           "pool",
	NodeTypePoolSupervisor: "pool-supervisor",
	NodeTypeSupervisor:     "supervisor",
}

type node struct {
	ID        uint64
	name      string
	nodeType  NodeType
	factory   func() Node
	startTime time.Time
	ctx       context.Context
}

var nodeCounter uint64

func generateNodeID() uint64 {
	return atomic.AddUint64(&nodeCounter, 1)
}

func newNode(ctx context.Context, name string, node_type NodeType, factory func() Node) node {
	n := node{
		ID:        generateNodeID(),
		nodeType:  node_type,
		startTime: time.Now(),
		factory:   factory,
		ctx:       ctx,
	}
	n.name = fmt.Sprintf("%s_%s_%d", nodeTypeNames[node_type], name, n.ID)
	return n
}

func (n node) GetID() uint64 {
	return n.ID
}

func (n node) GetName() string {
	return n.name
}

func (n node) Start() error {
	return nil
}

func (n node) Stop() error {
	return nil
}
