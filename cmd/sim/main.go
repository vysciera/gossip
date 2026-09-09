package main

import (
	"fmt"
	"time"

	"smalltalk/internal/hashgraph"
	"smalltalk/internal/network"
	"smalltalk/internal/sim"
)

func main() {
	// basicGossipDemo()
	// forkDemo()
	// stronglySeeDemo()
	// roundDemo()

	deterministicSimulationDemo()
}

//
// Basic gossip / hashgraph growth
//

func basicGossipDemo() {
	fmt.Println("\n--- basic gossip demo ---")

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

	fmt.Println("\ninitial")
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
}

func printState(nodes []*network.Node) {
	for _, node := range nodes {
		fmt.Printf(
			"\n%s [%s]: head=%s events=%d\n",
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

//
// Byzantine fork demo
//

func forkDemo() {
	fmt.Println("\n--- fork demo ---")

	alice := network.NewNode("alice")
	bob := network.NewNode("bob")
	carol := network.NewNode("carol")

	//
	// Alice's genesis event.
	//

	a0 := alice.Head

	//
	// Bob -> Alice
	//
	// Alice creates:
	//
	//        A1
	//       /  \
	//     A0    B0
	//

	if err := network.Gossip(bob, alice); err != nil {
		panic(err)
	}

	a1 := alice.Head

	//
	// Carol -> Alice
	//
	// Alice now learns C0 as well.
	//

	if err := network.Gossip(carol, alice); err != nil {
		panic(err)
	}

	c0 := carol.Head

	//
	// Alice cheats.
	//
	// Instead of extending her current self-chain,
	// she creates another index-1 event from A0.
	//
	//        A1
	//       /
	// A0 --<
	//       \
	//        A1'
	//

	fork := hashgraph.NewEvent(
		alice.PrivateKey,
		1,
		time.Now().UnixNano(),
		&a0,
		&c0,
	)

	if err := alice.Graph.Add(fork); err != nil {
		panic(err)
	}

	fmt.Printf(
		"A1  = %s\n",
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

	//
	// Alice tells Bob only about A1.
	//

	if err := network.GossipHead(
		alice,
		bob,
		a1,
	); err != nil {
		panic(err)
	}

	//
	// Alice tells Carol only about A1'.
	//

	if err := network.GossipHead(
		alice,
		carol,
		fork.ID,
	); err != nil {
		panic(err)
	}

	fmt.Printf(
		"bob detects fork?   %v\n",
		bob.Graph.IsFork(a1, fork.ID),
	)

	fmt.Printf(
		"carol detects fork? %v\n",
		carol.Graph.IsFork(a1, fork.ID),
	)

	//
	// Carol now tells Bob her history.
	//
	// Bob finally obtains both branches.
	//

	if err := network.Gossip(carol, bob); err != nil {
		panic(err)
	}

	fmt.Printf(
		"after carol -> bob, bob detects fork? %v\n",
		bob.Graph.IsFork(a1, fork.ID),
	)

	fmt.Println("\nseeing")

	fmt.Printf(
		"bob head has A1 as ancestor?  %v\n",
		bob.Graph.IsAncestor(a1, bob.Head),
	)

	fmt.Printf(
		"bob head sees A1?             %v\n",
		bob.Graph.See(bob.Head, a1),
	)

	fmt.Printf(
		"bob head has A1' as ancestor? %v\n",
		bob.Graph.IsAncestor(fork.ID, bob.Head),
	)

	fmt.Printf(
		"bob head sees A1'?            %v\n",
		bob.Graph.See(bob.Head, fork.ID),
	)
}

//
// Strongly-see demo
//

func stronglySeeDemo() {
	fmt.Println("\n--- strongly-see demo ---")

	alice := network.NewNode("alice")
	bob := network.NewNode("bob")
	carol := network.NewNode("carol")
	dave := network.NewNode("dave")

	membership := hashgraph.NewMembership(
		alice.ID,
		bob.ID,
		carol.ID,
		dave.ID,
	)

	a0 := alice.Head

	//
	// Spread A0 to three other members.
	//

	if err := network.Gossip(alice, bob); err != nil {
		panic(err)
	}

	if err := network.Gossip(alice, carol); err != nil {
		panic(err)
	}

	if err := network.Gossip(alice, dave); err != nil {
		panic(err)
	}

	//
	// Bob now learns what Carol and Dave know.
	//

	if err := network.Gossip(carol, bob); err != nil {
		panic(err)
	}

	if err := network.Gossip(dave, bob); err != nil {
		panic(err)
	}

	fmt.Printf(
		"bob head sees A0?          %v\n",
		bob.Graph.See(
			bob.Head,
			a0,
		),
	)

	fmt.Printf(
		"bob head strongly sees A0? %v\n",
		bob.Graph.StronglySee(
			bob.Head,
			a0,
			membership,
		),
	)
}

//
// Full round / witness / virtual voting /
// consensus ordering demo
//

func roundDemo() {
	fmt.Println("\n--- round / consensus demo ---")

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

	//
	// Save an early ordinary event so we can
	// inspect its later consensus metadata.
	//

	a0 := alice.Head

	membership := hashgraph.NewMembership(
		alice.ID,
		bob.ID,
		carol.ID,
		dave.ID,
	)

	type exchange struct {
		from *network.Node
		to   *network.Node
	}

	//
	// Deliberately interconnected gossip schedule.
	//

	schedule := []exchange{
		{alice, bob},
		{bob, carol},
		{carol, dave},
		{dave, alice},

		{alice, carol},
		{carol, bob},
		{bob, dave},
		{dave, carol},

		{carol, alice},
		{alice, dave},
		{dave, bob},
		{bob, alice},
	}

	//
	// Generate enough history for several rounds,
	// fame elections, and finalized events.
	//

	for cycle := 0; cycle < 10; cycle++ {
		for _, exchange := range schedule {
			if err := network.Gossip(
				exchange.from,
				exchange.to,
			); err != nil {
				panic(err)
			}
		}
	}

	//
	// Inspect derived rounds.
	//

	for _, node := range nodes {
		printRounds(
			node,
			membership,
		)
	}

	//
	// Inspect Alice's first-round witnesses
	// and one fame election.
	//

	witnesses := alice.Graph.WitnessesByRound(
		alice.Head,
		membership,
	)

	if len(witnesses[1]) > 0 {
		candidate := witnesses[1][0]

		printFirstFameVotes(
			alice,
			membership,
		)

		printFameElection(
			alice,
			membership,
			candidate,
		)

		printFameDecision(
			alice,
			membership,
			candidate,
		)
	}

	//
	// Inspect A0's consensus metadata.
	//

	printRoundReceived(
		alice,
		membership,
		a0,
	)

	printConsensusTimestamp(
		alice,
		membership,
		a0,
	)

	//
	// Finally compare the consensus order each
	// independent node derives.
	//

	for _, node := range nodes {
		printConsensusOrder(
			node,
			membership,
		)
	}
}

func printRounds(
	node *network.Node,
	membership *hashgraph.Membership,
) {
	info := node.Graph.DivideRounds(
		node.Head,
		membership,
	)

	fmt.Printf(
		"\n%s rounds\n",
		node.Name,
	)

	for _, event := range node.Graph.History(node.Head) {
		roundInfo, ok := info[event.ID]

		if !ok {
			continue
		}

		witness := ""

		if roundInfo.Witness {
			witness = " WITNESS"
		}

		fmt.Printf(
			"  creator=%s index=%-3d round=%d%s\n",
			event.Creator.Short(),
			event.Index,
			roundInfo.Round,
			witness,
		)
	}
}

func printFirstFameVotes(
	node *network.Node,
	membership *hashgraph.Membership,
) {
	info := node.Graph.DivideRounds(
		node.Head,
		membership,
	)

	witnesses := node.Graph.WitnessesByRound(
		node.Head,
		membership,
	)

	fmt.Printf(
		"\n%s first virtual votes\n",
		node.Name,
	)

	for round, candidates := range witnesses {
		voters := witnesses[round+1]

		if len(voters) == 0 {
			continue
		}

		fmt.Printf(
			"\nround %d candidates\n",
			round,
		)

		for _, candidate := range candidates {
			fmt.Printf(
				"  candidate %s\n",
				candidate.Short(),
			)

			for _, voter := range voters {
				vote, ok :=
					node.Graph.FirstFameVote(
						voter,
						candidate,
						info,
					)

				if !ok {
					continue
				}

				voterEvent, ok :=
					node.Graph.Get(voter)

				if !ok {
					continue
				}

				fmt.Printf(
					"    creator=%s voter=%s -> %s\n",
					voterEvent.Creator.Short(),
					voter.Short(),
					vote,
				)
			}
		}
	}
}

func printFameElection(
	node *network.Node,
	membership *hashgraph.Membership,
	candidate hashgraph.EventID,
) {
	info := node.Graph.DivideRounds(
		node.Head,
		membership,
	)

	witnesses := node.Graph.WitnessesByRound(
		node.Head,
		membership,
	)

	votes := node.Graph.FameVotes(
		node.Head,
		candidate,
		membership,
	)

	candidateInfo, ok := info[candidate]

	if !ok {
		return
	}

	fmt.Printf(
		"\nfame election for %s (round %d)\n",
		candidate.Short(),
		candidateInfo.Round,
	)

	for round := candidateInfo.Round + 1; ; round++ {
		roundWitnesses, ok :=
			witnesses[round]

		if !ok {
			break
		}

		fmt.Printf(
			"\nround %d\n",
			round,
		)

		for _, voter := range roundWitnesses {
			vote, ok := votes[hashgraph.VoteKey{
				Candidate: candidate,
				Voter:     voter,
			}]

			if !ok {
				continue
			}

			event, ok :=
				node.Graph.Get(voter)

			if !ok {
				continue
			}

			fmt.Printf(
				"  %s [%s] -> %s\n",
				event.Creator.Short(),
				voter.Short(),
				vote,
			)
		}
	}
}

func printFameDecision(
	node *network.Node,
	membership *hashgraph.Membership,
	candidate hashgraph.EventID,
) {
	result := node.Graph.DecideFame(
		node.Head,
		candidate,
		membership,
		hashgraph.DefaultCoinPeriod,
	)

	fmt.Printf(
		"\nfame decision for %s\n",
		candidate.Short(),
	)

	fmt.Printf(
		"  result: %s\n",
		result.Fame,
	)

	if result.Fame == hashgraph.FameUndecided {
		return
	}

	fmt.Printf(
		"  decision round: %d\n",
		result.DecisionRound,
	)

	fmt.Printf(
		"  deciding witness: %s\n",
		result.Decider.Short(),
	)
}

func printRoundReceived(
	node *network.Node,
	membership *hashgraph.Membership,
	eventID hashgraph.EventID,
) {
	roundReceived, ok :=
		node.Graph.RoundReceived(
			node.Head,
			eventID,
			membership,
		)

	fmt.Printf(
		"\nround received for %s\n",
		eventID.Short(),
	)

	if !ok {
		fmt.Println("  UNDECIDED")
		return
	}

	fmt.Printf(
		"  round %d\n",
		roundReceived,
	)
}

func printConsensusTimestamp(
	node *network.Node,
	membership *hashgraph.Membership,
	eventID hashgraph.EventID,
) {
	timestamp, ok :=
		node.Graph.ConsensusTimestamp(
			node.Head,
			eventID,
			membership,
		)

	fmt.Printf(
		"\nconsensus timestamp for %s\n",
		eventID.Short(),
	)

	if !ok {
		fmt.Println("  UNDECIDED")
		return
	}

	fmt.Printf(
		"  %s\n",
		time.Unix(
			0,
			timestamp,
		).Format(time.RFC3339Nano),
	)
}

func printConsensusOrder(
	node *network.Node,
	membership *hashgraph.Membership,
) {
	ordered := node.Graph.ConsensusOrder(
		node.Head,
		membership,
	)

	fmt.Printf(
		"\n%s consensus order\n\n",
		node.Name,
	)

	if len(ordered) == 0 {
		fmt.Println("  no consensus events yet")
		return
	}

	for i, item := range ordered {
		fmt.Printf(
			"  %03d  creator=%s index=%-3d round=%-3d time=%s event=%s\n",
			i,
			item.Event.Creator.Short(),
			item.Event.Index,
			item.RoundReceived,
			time.Unix(
				0,
				item.ConsensusTimestamp,
			).Format("15:04:05.000000"),
			item.Event.ID.Short(),
		)
	}
}

func deterministicSimulationDemo() {
	fmt.Println(
		"\n--- deterministic simulation ---",
	)

	s := sim.New(
		42,
		"alice",
		"bob",
		"carol",
		"dave",
	)

	if err := s.Run(100); err != nil {
		panic(err)
	}

	for _, node := range s.Nodes {
		order := node.Graph.ConsensusOrder(node.Head, s.Membership)

		fmt.Printf(
			"%s: events=%d consensus=%d head=%s\n",
			node.Name,
			node.Graph.Len(),
			len(order),
			node.Head.Short(),
		)
	}

	common, agreement := s.ConsensusPrefixesAgree()

	fmt.Printf("\ncommon consensus prefix: %d\n", common)
	fmt.Printf("agreement: %v\n", agreement)
}
