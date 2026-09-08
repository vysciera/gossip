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

	nodes := []*network.Node{
		alice,
		bob,
		carol,
		dave,
	}

	fmt.Println("initial")
	printState(nodes)

	fmt.Println("\nalice -> bob")
	if err := network.Gossip(alice, bob); err != nil {
		panic(err)
	}

	fmt.Println("\nbob -> carol")
	if err := network.Gossip(bob, carol); err != nil {
		panic(err)
	}

	fmt.Println("\ndave -> alice")
	if err := network.Gossip(dave, alice); err != nil {
		panic(err)
	}

	fmt.Println("\ncarol -> dave")
	if err := network.Gossip(carol, dave); err != nil {
		panic(err)
	}

	fmt.Println("\nfinal")
	printState(nodes)

	forkDemo()
}

func printState(nodes []*network.Node) {
	for _, node := range nodes {
		fmt.Printf(
			"\n%s: [%s]: head=%s events=%d\n",
			node.Name,
			node.ID.Short(),
			node.Head.Short(),
			node.Graph.Len(),
		)

		printKnowledge(node)
	}
}

func printKnowledge(node *network.Node) {
	events := node.Graph.Ancestors(node.Head)

	for _, event := range events {
		fmt.Printf(
			"  creator=%s index=%d event=%s\n",
			event.Creator.Short(),
			event.Index,
			event.ID.Short(),
		)
	}
}

func forkDemo() {
	fmt.Println("\n--- fork demo ---")

	alice := network.NewNode("alice")
	bob := network.NewNode("bob")
	carol := network.NewNode("carol")

	// Save Alice's initial event
	a0 := alice.Head

	// Alive learns Bob's history normally.
	// This creates:
	//
	//     A1
	//    /  \
	//  A0    B0

	if err := network.Gossip(bob, alice); err != nil {
		panic(err)
	}

	a1 := alice.Head

	// Alice learns Carol's history
	// Normally, Alice continues: A0 -> A1 -> A2

	if err := network.Gossip(carol, alice); err != nil {
		panic(err)
	}

	// Alice now knows C0
	c0 := carol.Head

	// Alice cheats. Instead of building on A1/A2, she goes back to A0
	// and creates another event at index 1:
	//
	//				  A1
	//		A0 ----<
	//				  A1'

	fork := hashgraph.NewEvent(
		alice.PrivateKey,
		1,
		&a0,
		&c0,
	)

	if err := alice.Graph.Add(fork); err != nil {
		panic(err)
	}

	fmt.Printf(
		"A1 = %s\n",
		a1.Short(),
	)

	fmt.Printf(
		"A1' = %s\n",
		fork.ID.Short(),
	)

	fmt.Printf(
		"fork? %v\n",
		alice.Graph.IsFork(a1, fork.ID),
	)

	// Alice presents different histories to Bob and Carol
	if err := network.GossipHead(
		alice,
		bob,
		a1,
	); err != nil {
		panic(err)
	}

	if err := network.GossipHead(
		alice,
		carol,
		fork.ID,
	); err != nil {
		panic(err)
	}

	fmt.Printf(
		"bob detects fork?	%v\n",
		bob.Graph.IsFork(a1, fork.ID),
	)

	fmt.Printf(
		"carol detects fork?	%v\n",
		carol.Graph.IsFork(a1, fork.ID),
	)

	// Carol now gossips what she knows to Bob
	if err := network.Gossip(carol, bob); err != nil {
		panic(err)
	}

	fmt.Println("\nseeing")

	fmt.Printf(
		"bob head has A1 as ancestor?	%v\n",
		bob.Graph.IsAncestor(a1, bob.Head),
	)

	fmt.Printf(
		"bob head sees A1?				%v\n",
		bob.Graph.See(bob.Head, a1),
	)

	fmt.Printf(
		"bob head has A1' as ancestor	%v\n",
		bob.Graph.IsAncestor(fork.ID, bob.Head),
	)

	fmt.Printf(
		"bob head sees A1'?				%v\n",
		bob.Graph.See(bob.Head, fork.ID),
	)

	fmt.Printf(
		"after carol -> bob, bob detects fork?	%v\n",
		bob.Graph.IsFork(a1, fork.ID),
	)

}
