package network

import (
	"smalltalk/internal/hashgraph"
)

/*

Unidirectional microgossip:
A, B = {X, Y, Z}, {A, B}
Gossip(A, B) -> A, B: {X, Y, Z}, {A, B, X, Y, Z}

Events create new states of knowledge.

*/

func Gossip(from, to *Node) {
	for value := range from.Known {
		to.Learn(value)
	}

	to.Head = &hashgraph.Event{
		Creator:		to.Name,
		SelfParent:		to.Head,
		OtherParent:	from.Head,
	}
}

