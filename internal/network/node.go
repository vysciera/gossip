package network

import (
	"smalltalk/internal/hashgraph"
)

// Node state records:
// set of knowledge + a history describing how knowledge arrived to the node

type Node struct {
	Name		string
	Head		*hashgraph.Event
	NextIndex	int
}

func NewNode(name string) *Node {
	n := &Node{Name: name, NextIndex: 1}
	n.Head = &hashgraph.Event{Creator: name, Index: 0}
		
	return n
}

