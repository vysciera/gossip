package hashgraph

type Event struct {
	Creator	string

	SelfParent	*Event
	OtherParent	*Event
}
