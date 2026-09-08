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
	to.Head = &hashgraph.Event{
		Creator:		to.Name,
		Index:			to.NextIndex,
		SelfParent:		to.Head,
		OtherParent:	from.Head,
	}

	to.NextIndex++
}
