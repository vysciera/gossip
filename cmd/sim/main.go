package main

import (
	"fmt"

	"smalltalk/internal/hashgraph"
	"smalltalk/internal/network"
)

func main() {
	alice := network.NewNode("alice")
	bob := network.NewNode("bob")
	carol := network.NewNode("carol")
	dave := network.NewNode("dave")

	fmt.Println("initial state")
	printKnowledge(alice)
	printKnowledge(bob)
	printKnowledge(carol)
	printKnowledge(dave)

	fmt.Println("\nalice -> bob")
	network.Gossip(alice, bob)

	fmt.Println("\nbob -> carol")
	network.Gossip(bob, carol)

	fmt.Println("\ndave -> alice")
	network.Gossip(dave, alice)

	fmt.Println("\ncarol -> dave")
	network.Gossip(carol, dave)

	fmt.Println("\nfinal state")
	printKnowledge(alice)
	printKnowledge(bob)
	printKnowledge(carol)
	printKnowledge(dave)
}

func printKnowledge(node *network.Node) {
	events := hashgraph.Ancestors(node.Head)

	fmt.Printf("\n%s knows:\n", node.Name)

	for _, event := range events {
		fmt.Printf(
			"  %s%d\n",
			event.Creator,
			event.Index,
		)
	}
}
