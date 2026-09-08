package hashgraph

type RoundInfo struct {
	Round		uint64
	Witness		bool
}

func (g *Graph) DivideRounds(head EventID, membership *Membership) map[EventID]RoundInfo {
	info := make(map[EventID]RoundInfo)

	if membership == nil || membership.Len() == 0 {
		return info
	}

	// History gives us parents before children.
	// When we calculate an event's round, both parent rounds are already known.
	history := g.History(head)

	witnesses := make(map[uint64][]EventID)

	for _, event := range history {
		round := uint64(1)

		// Genesis event.
		if event.SelfParent == nil && event.OtherParent == nil {
			info[event.ID] = RoundInfo{
				Round:		1,
				Witness:	true,
			}

			witnesses[1] = append(witnesses[1], event.ID)

			continue
		}

		// Start at the greater parent round.
		selfInfo := info[*event.SelfParent]
		otherInfo := info[*event.OtherParent]

		round = max(selfInfo.Round, otherInfo.Round)

		// Does this event strongly see a supermajority
		// of witnesses from that round?
		count := 0

		for _, witnessID := range witnesses[round] {
			if g.StronglySee(
				event.ID,
				witnessID,
				membership,
			) {
				count++
			}
		}

		if count * 3 > membership.Len() * 2 {
			round++
		}

		// First event by this creator in the new round?
		isWitness := round > selfInfo.Round

		info[event.ID] = RoundInfo{
			Round:		round,
			Witness:	isWitness,
		}

		if isWitness {
			witnesses[round] = append(witnesses[round], event.ID)
		}
	}

	return info
}
