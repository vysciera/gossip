package network

/* Unidirectional microgossip.

Alice knows: {X, Y, Z}
Bob knows:	 {A, B}

Gossip(alice, bob) ->

Alice: {X, Y, Z}
Bob:   {A, B, X, Y, Z} */
func Gossip(from, to *Node) {
	for value := range from.Known {
		to.Learn(value)
	}
}

