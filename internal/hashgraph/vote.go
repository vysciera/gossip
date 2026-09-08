package hashgraph

type Vote bool

const (
	VoteNo  Vote = false
	VoteYes Vote = true
)

type VoteTally struct {
	Yes int
	No  int
}

type VoteKey struct {
	Candidate EventID
	Voter     EventID
}

func (t VoteTally) Majority() Vote {
	// Hashgraph breaks ties in favor of YES
	if t.Yes >= t.No {
		return VoteYes
	}

	return VoteNo
}

func (t VoteTally) Count(vote Vote) int {
	if vote == VoteYes {
		return t.Yes
	}

	return t.No
}

func (v Vote) String() string {
	if v {
		return "YES"
	}

	return "NO"
}

func (g *Graph) FirstFameVote(voter EventID, candidate EventID, info map[EventID]RoundInfo) (Vote, bool) {
	voterInfo, voterOK := info[voter]
	candidateInfo, candidateOK := info[candidate]

	if !voterOK || !candidateOK {
		return VoteNo, false
	}

	if !voterInfo.Witness || !candidateInfo.Witness {
		return VoteNo, false
	}

	if voterInfo.Round != candidateInfo.Round+1 {
		return VoteNo, false
	}

	if g.See(voter, candidate) {
		return VoteYes, true
	}

	return VoteNo, true
}

func (g *Graph) LaterFameVote(
	voter EventID,
	candidate EventID,
	info map[EventID]RoundInfo,
	witnesses map[uint64][]EventID,
	votes map[VoteKey]Vote,
	membership *Membership,
) (Vote, VoteTally, bool) {
	voterInfo, voterOK := info[voter]
	candidateInfo, candidateOK := info[candidate]

	if !voterOK || !candidateOK {
		return VoteNo, VoteTally{}, false
	}

	if !voterInfo.Witness || !candidateInfo.Witness {
		return VoteNo, VoteTally{}, false
	}

	// Handles only the recursive part of an election
	if voterInfo.Round <= candidateInfo.Round+1 {
		return VoteNo, VoteTally{}, false
	}

	previousRound := voterInfo.Round - 1

	var tally VoteTally

	for _, previousWitness := range witnesses[previousRound] {
		// A voter only considers previous-round
		// witnesses that it strongly sees.
		if !g.StronglySee(voter, previousWitness, membership) {
			continue
		}

		previousVote, ok := votes[VoteKey{
			Candidate: candidate,
			Voter:     previousWitness,
		}]

		if !ok {
			continue
		}

		if previousVote == VoteYes {
			tally.Yes++
		} else {
			tally.No++
		}
	}

	return tally.Majority(), tally, true
}

func (g *Graph) FameVotes(head EventID, candidate EventID, membership *Membership) map[VoteKey]Vote {
	votes := make(map[VoteKey]Vote)
	info := g.DivideRounds(head, membership)

	candidateInfo, ok := info[candidate]
	if !ok || !candidateInfo.Witness {
		return votes
	}

	witnesses := g.WitnessesByRound(head, membership)

	// First voting round
	firstRound := candidateInfo.Round + 1

	for _, voter := range witnesses[firstRound] {
		vote, ok := g.FirstFameVote(voter, candidate, info)

		if !ok {
			continue
		}

		votes[VoteKey{
			Candidate: candidate,
			Voter:     voter,
		}] = vote
	}

	// Later rounds.
	for round := firstRound + 1; ; round++ {
		voters, exists := witnesses[round]
		if !exists {
			break
		}

		for _, voter := range voters {
			vote, _, ok := g.LaterFameVote(
				voter,
				candidate,
				info,
				witnesses,
				votes,
				membership,
			)

			if !ok {
				continue
			}

			votes[VoteKey{
				Candidate: candidate,
				Voter:     voter,
			}] = vote
		}
	}

	return votes
}
