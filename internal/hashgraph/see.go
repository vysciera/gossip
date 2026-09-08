package hashgraph

func (g *Graph) See(x, y EventID) bool {
	if !g.Has(x) {
		return false
	}

	yEvent, ok := g.Get(y)
	if !ok {
		return false
	}

	if !g.IsAncestor(y, x) {
		return false
	}

	if g.HasForkBy(x, yEvent.Creator) {
		return false
	}

	return true
}

func (g *Graph) StronglySee(x EventID, y EventID, membership *Membership) bool {
	if membership == nil || membership.Len() == 0 {
		return false
	}

	if !g.Has(x) || !g.Has(y) {
		return false
	}

	seenCreators := make(map[NodeID]struct{})

	for _, event := range g.Ancestors(x) {
		if !membership.Contains(event.Creator) {
			continue
		}

		// x must be able to see this intermediary event.
		if !g.See(x, event.ID) {
			continue
		}

		// the intermediary event must itself see y.
		if !g.See(event.ID, y) {
			continue
		}

		seenCreators[event.Creator] = struct{}{}
	}

	count := len(seenCreators)
	n := membership.Len()

	// count > 2n/3
	return count*3 > n*2
}
