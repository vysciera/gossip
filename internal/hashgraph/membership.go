package hashgraph

type Membership struct {
	members map[NodeID]struct{}
}

func NewMembership(ids ...NodeID) *Membership {
	members := make(map[NodeID]struct{}, len(ids))

	for _, id := range ids {
		members[id] = struct{}{}
	}

	return &Membership{
		members: members,
	}
}

func (m *Membership) Contains(id NodeID) bool {
	_, ok := m.members[id]
	return ok
}

func (m *Membership) Len() int {
	return len(m.members)
}

func (m *Membership) IsSupermajority(count int) bool {
	if m == nil || len(m.members) == 0 {
		return false
	}

	return count*3 > len(m.members)*2
}
