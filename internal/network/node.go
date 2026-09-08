package network

type Node struct {
	Name	string
	Known	map[string]struct{}
}

func NewNode(name string) *Node {
	n := &Node{
		Name:	name,
		Known:	make(map[string]struct{}),
	}

	// Each node initializes knowing only its own name
	n.Known[name] = struct{}{}

	return n
}

func (n *Node) Learn(value string) {
	n.Known[value] = struct{}{}
}

func (n *Node) Knows(value string) bool {
	_, ok := n.Known[value]
	return ok
}

