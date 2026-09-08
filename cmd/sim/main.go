package main

import (
	"fmt"
	"sort"

	"smalltalk/internal/network"
	"smalltalk/internal/hashgraph"
)

func main() {
	alice := network.NewNode("alice")
	bob := network.NewNode("bob")
	carol := network.NewNode("carol")
	dave := network.NewNode("dave")

	a0 := alice.Head
	b0 := bob.Head
	c0 := carol.Head
	d0 := dave.Head

	nodes := []*network.Node{
		alice,
		bob,
		carol,
		dave,
	}

	printState("initial", nodes)

	gossip("alice -> bob", alice, bob, nodes)
	gossip("bob -> carol", bob, carol, nodes)
	gossip("dave -> alice", dave, alice, nodes)
	gossip("carol -> dave", carol, dave, nodes)

	fmt.Println("\nancestry:")
	fmt.Printf(
		"does dave know alice's original event? %v\n",
		hashgraph.IsAncestor(a0, dave.Head),
	)

	fmt.Printf(
		"does dave know bob's original event? %v\n",
		hashgraph.IsAncestor(b0, dave.Head),
	)

	fmt.Printf(
		"does dave know carol's original event? %v\n",
		hashgraph.IsAncestor(c0, dave.Head),
	)

	fmt.Printf(
		"did alice learn dave's original event? %v\n",
		hashgraph.IsAncestor(d0, alice.Head),
	)	
}

func gossip(label string, from *network.Node, to *network.Node, nodes []*network.Node) {
	network.Gossip(from, to)
	printState(label, nodes)
}

func printState(label string, nodes []*network.Node) {
	fmt.Printf("\n%s\n\n", label)

	for _, node := range nodes {
		values := make([]string, 0, len(node.Known))

		for value := range node.Known {
			values = append(values, value)
		}

		sort.Strings(values)

		// Evil
		fmt.Printf("%-6s %v\n", node.Name, values)
	}
}
