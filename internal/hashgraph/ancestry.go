package hashgraph

func (g *Graph) IsAncestor(ancestor, descendant EventID) bool {
	if !g.Has(ancestor) || !g.Has(descendant) {
		return false
	}

	stack := []EventID{descendant}
	visited := make(map[EventID]struct{})

	for len(stack) > 0 {
		last := len(stack) - 1

		currentID := stack[last]
		stack = stack[:last]

		if currentID == ancestor {
			return true
		}

		if _, seen := visited[currentID]; seen {
			continue
		}

		visited[currentID] = struct{}{}

		current, ok := g.Get(currentID)
		if !ok {
			continue
		}

		if current.SelfParent != nil {
			stack = append(stack, *current.SelfParent)
		}

		if current.OtherParent != nil {
			stack = append(stack, *current.OtherParent)
		}
	}

	return false
}

func (g *Graph) Ancestors(id EventID) []Event {
	if !g.Has(id) {
		return nil
	}

	stack := []EventID{id}
	visited := make(map[EventID]struct{})

	var result []Event

	for len(stack) > 0 {
		last := len(stack) - 1

		currentID := stack[last]
		stack = stack[:last]

		if _, seen := visited[currentID]; seen {
			continue
		}

		visited[currentID] = struct{}{}

		current, ok := g.Get(currentID)
		if !ok {
			continue
		}

		result = append(result, current)

		if current.SelfParent != nil {
			stack = append(stack, *current.SelfParent)
		}

		if current.OtherParent != nil {
			stack = append(stack, *current.OtherParent)
		}
	}

	return result
}

func (g *Graph) History(id EventID) []Event {
	if !g.Has(id) {
		return nil
	}

	visited := make(map[EventID]struct{})
	var result []Event

	var walk func(EventID)

	walk = func(currentID EventID) {
		if _, seen := visited[currentID]; seen {
			return
		}

		visited[currentID] = struct{}{}

		current, ok := g.Get(currentID)
		if !ok {
			return
		}

		if current.SelfParent != nil {
			walk(*current.SelfParent)
		}

		if current.OtherParent != nil {
			walk(*current.OtherParent)
		}

		result = append(result, current)
	}

	walk(id)

	return result
}

func (g *Graph) IsSelfAncestor(ancestor EventID, descendant EventID) bool {
	if !g.Has(ancestor) || !g.Has(descendant) {
		return false
	}

	current := descendant
	visited := make(map[EventID]struct{})

	for {
		if current == ancestor {
			return true
		}

		if _, seen := visited[current]; seen {
			return false
		}

		visited[current] = struct{}{}

		event, ok := g.Get(current)
		if !ok {
			return false
		}

		if event.SelfParent == nil {
			return false
		}

		current = *event.SelfParent
	}
}

func (g *Graph) IsFork(left, right EventID) bool {
	leftEvent, leftOK := g.Get(left)
	rightEvent, rightOK := g.Get(right)

	if !leftOK || !rightOK {
		return false
	}

	if leftEvent.Creator != rightEvent.Creator {
		return false
	}

	return !g.IsSelfAncestor(left, right) && !g.IsSelfAncestor(right, left)
}
