package soop

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type Node interface {
	GetName() string
}

type NodeType int

const (
	NodeTypeDefault NodeType = iota
	NodeTypeWorker
	NodeTypePool
	NodeTypeSupervisor
)

var nodeTypeNames = map[NodeType]string{
	NodeTypeDefault:    "node",
	NodeTypeWorker:     "worker",
	NodeTypePool:       "pool",
	NodeTypeSupervisor: "supervisor",
}

type node struct {
	ID        uint64
	name      string
	typeName  NodeType
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
		typeName:  node_type,
		startTime: time.Now(),
		factory:   factory,
		ctx:       ctx,
	}
	n.name = fmt.Sprintf("%s_%s_%d", nodeTypeNames[node_type], name, n.ID)
	return n
}

func (n node) GetName() string {
	return n.name
}
