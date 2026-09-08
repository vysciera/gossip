package hashgraph

type Event struct {
	Creator	string
	Index	int // temp hash placeholder

	SelfParent	*Event
	OtherParent	*Event
}
