package hashgraph

type Vote bool
type Fame uint8

const DefaultCoinPeriod uint64 = 10

const (
	VoteNo  Vote = false
	VoteYes Vote = true
)

const (
	FameUndecided Fame = iota
	FameNotFamous
	FameFamous
)

type FameResult struct {
	Fame Fame

	DecisionRound uint64
	Decider       EventID

	Votes map[VoteKey]Vote
}

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

func (f Fame) String() string {
	switch f {
	case FameFamous:
		return "FAMOUS"

	case FameNotFamous:
		return "NOT FAMOUS"

	default:
		return "UNDECIDED"
	}
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

	if tally.Yes+tally.No == 0 {
		return VoteNo, tally, false
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

func (g *Graph) DecideFame(head EventID, candidate EventID, membership *Membership, coinPeriod uint64) FameResult {
	result := FameResult{
		Fame:  FameUndecided,
		Votes: make(map[VoteKey]Vote),
	}

	if membership == nil || membership.Len() == 0 {
		return result
	}

	if coinPeriod <= 2 {
		coinPeriod = DefaultCoinPeriod
	}

	info := g.DivideRounds(head, membership)
	candidateInfo, ok := info[candidate]

	if !ok || !candidateInfo.Witness {
		return result
	}

	witnesses := g.WitnessesByRound(head, membership)
	firstVotingRound := candidateInfo.Round + 1

	// First-round witnesses vote according to
	// whether they can see the candidate

	for _, voter := range witnesses[firstVotingRound] {

		vote, ok := g.FirstFameVote(voter, candidate, info)
		if !ok {
			continue
		}

		result.Votes[VoteKey{
			Candidate: candidate,
			Voter:     voter,
		}] = vote
	}

	// d >= 2
	for round := firstVotingRound + 1; ; round++ {
		voters, exists := witnesses[round]
		if !exists {
			break
		}

		d := round - candidateInfo.Round

		coinRound := d%coinPeriod == 0

		for _, voter := range voters {
			majorityVote, tally, ok := g.LaterFameVote(
				voter,
				candidate,
				info,
				witnesses,
				result.Votes,
				membership,
			)

			if !ok {
				continue
			}

			majorityCount := tally.Count(majorityVote)
			supermajority := membership.IsSupermajority(majorityCount)
			key := VoteKey{
				Candidate: candidate,
				Voter:     voter,
			}

			// Normal round
			if !coinRound {
				result.Votes[key] = majorityVote

				if !supermajority {
					continue
				}

				if majorityVote == VoteYes {
					result.Fame = FameFamous
				} else {
					result.Fame = FameNotFamous
				}

				result.DecisionRound = round
				result.Decider = voter

				return result
			}

			// Coin round
			// A supermajority is adopted, but
			// does NOT terminate the election.

			if supermajority {
				result.Votes[key] = majorityVote

				continue
			}

			// No supermajority:
			// Dervice the vote from the witness' signature.

			event, ok := g.Get(voter)

			if !ok {
				continue
			}

			result.Votes[key] = CoinVote(event)
		}
	}

	return result
}

func (g *Graph) DecideFameWithoutCoins(head EventID, candidate EventID, membership *Membership) FameResult {
	result := FameResult{
		Fame:  FameUndecided,
		Votes: make(map[VoteKey]Vote),
	}

	if membership == nil || membership.Len() == 0 {
		return result
	}

	info := g.DivideRounds(head, membership)

	candidateInfo, ok := info[candidate]
	if !ok || !candidateInfo.Witness {
		return result
	}

	witnesses := g.WitnessesByRound(head, membership)

	// Round :: r + 1
	// First Vote:
	//	Can voter see candidate?

	firstVotingRound := candidateInfo.Round + 1

	for _, voter := range witnesses[firstVotingRound] {
		vote, ok := g.FirstFameVote(voter, candidate, info)

		if !ok {
			continue
		}

		result.Votes[VoteKey{
			Candidate: candidate,
			Voter:     voter,
		}] = vote
	}

	// Round :: r + 2 (onward)
	// Later witnesses use the majority vote of
	// previous-round witnesses they strongly see.

	for round := firstVotingRound + 1; ; round++ {
		voters, exists := witnesses[round]

		if !exists {
			break
		}

		for _, voter := range voters {
			vote, tally, ok := g.LaterFameVote(
				voter,
				candidate,
				info,
				witnesses,
				result.Votes,
				membership,
			)

			if !ok {
				continue
			}

			result.Votes[VoteKey{
				Candidate: candidate,
				Voter:     voter,
			}] = vote

			majorityCount := tally.Count(vote)

			if !membership.IsSupermajority(majorityCount) {
				continue
			}

			if vote == VoteYes {
				result.Fame = FameFamous
			} else {
				result.Fame = FameNotFamous
			}

			result.DecisionRound = round
			result.Decider = voter

			return result
		}
	}

	return result
}

func CoinVote(event Event) Vote {
	if len(event.Signature) == 0 {
		return VoteNo
	}

	// Interpret signature as a sequence of bits, nibble
	bitIndex := len(event.Signature) * 8 / 2

	byteIndex := bitIndex / 8
	offset := uint(bitIndex % 8)

	bit := (event.Signature[byteIndex] >> offset) & 1

	if bit == 1 {
		return VoteYes
	}

	return VoteNo
}
