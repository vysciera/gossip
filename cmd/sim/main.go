package main

import (
	"fmt"
	"sort"

	"smalltalk/internal/network"
)

func main() {
	alice := network.NewNode("alice")
	bob := network.NewNode("bob")
	carol := network.NewNode("carol")
	dave := network.NewNode("dave")

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
