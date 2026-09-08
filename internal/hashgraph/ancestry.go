package hashgraph

// Starting at D1, can I repeatedly follow parent links backwards until I reach A0?
// Iterative, depth-first

func IsAncestor(ancestor, descendant *Event) bool {
	if ancestor == nil || descendant == nil {
		return false
	}

	stack := []*Event{descendant}
	visited := make(map[*Event]struct{})

	for len(stack) > 0 {

		// Pop last event off stack 
		last := len(stack) - 1

		current := stack[last]
		stack = stack[:last]

		if current == ancestor {
			return true
		}

		if _, seen := visited[current]; seen {
			continue
		}

		visited[current] = struct{}{}

		if current.SelfParent != nil {
			stack = append(stack, current.SelfParent)
		}

		if current.OtherParent != nil {
			stack = append(stack, current.OtherParent)
		}
	}

	return false
}

func Ancestors(event *Event) []*Event {
	if event == nil {
		return nil
	}

	stack := []*Event{event}
	visited := make(map[*Event]struct{})

	var result []*Event

	for len(stack) > 0 {
		last := len(stack) - 1

		current := stack[last]
		stack = stack[:last]

		if _, seen := visited[current]; seen {
			continue
		}

		visited[current] = struct{}{}
		result = append(result, current)

		if current.SelfParent != nil {
			stack = append(stack, current.SelfParent)
		}

		if current.OtherParent != nil {
			stack = append(stack, current.OtherParent)
		}
	}

	return result
}
