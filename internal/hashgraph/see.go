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
