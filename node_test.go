package soop

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateNodeID(t *testing.T) {
	initialID := nodeCounter
	id1 := generateNodeID()
	id2 := generateNodeID()

	assert.Equal(t, initialID+1, id1, "Expected node ID to increment by 1")
	assert.Equal(t, initialID+2, id2, "Expected node ID to increment by 1 again")
	assert.NotEqual(t, id1, id2, "Expected generated node IDs to be unique")
}

func TestNewNode(t *testing.T) {
	ctx := context.Background()
	name := "test"
	nodeType := NodeTypeWorker
	factory := func() Node { return &node{} }

	n := newNode(ctx, name, nodeType, factory)

	assert.NotNil(t, n, "Expected new node to be non-nil")
	assert.Equal(t, nodeType, n.nodeType, "Expected node type to be set correctly")
	assert.Equal(t, ctx, n.ctx, "Expected context to be set correctly")
	assert.NotEmpty(t, n.name, "Expected node name to be non-empty")
	assert.Contains(t, n.name, nodeTypeNames[nodeType], "Expected node name to contain node type name")
	assert.Contains(t, n.name, name, "Expected node name to contain the provided name")
	assert.Greater(t, n.ID, uint64(0), "Expected node ID to be greater than zero")
	assert.WithinDuration(t, time.Now(), n.startTime, time.Second, "Expected node start time to be set to now")
}

func TestNodeGetName(t *testing.T) {
	ctx := context.Background()
	name := "test"
	nodeType := NodeTypeWorker
	factory := func() Node { return &node{} }

	n := newNode(ctx, name, nodeType, factory)

	assert.Equal(t, n.name, n.GetName(), "Expected GetName to return the node's name")
}
