package hashgraph

import (
	"bytes"
	"sort"
)

type ConsensusEvent struct {
	Event Event

	RoundReceived      uint64
	ConsensusTimestamp int64

	WhitenedSignature []byte
}

// FameDecisions calculates the fame election results
// for every witness currently known beneath head.
func (g *Graph) FameDecisions(head EventID, membership *Membership) map[EventID]FameResult {
	results := make(map[EventID]FameResult)
	witnesses := g.WitnessesByRound(head, membership)

	for _, roundWitnesses := range witnesses {
		for _, witness := range roundWitnesses {
			results[witness] = g.DecideFame(
				head,
				witness,
				membership,
				DefaultCoinPeriod,
			)
		}
	}

	return results
}

// AllFameDecidedThrough reports whether every witness
// from round 1 through round r has a final fame decision.
func AllFameDecidedThrough(r uint64, witnesses map[uint64][]EventID, fame map[EventID]FameResult) bool {
	for round := uint64(1); round <= r; round++ {
		for _, witness := range witnesses[round] {
			result, ok := fame[witness]

			if !ok {
				return false
			}

			if result.Fame == FameUndecided {
				return false
			}
		}
	}

	return false
}

// UniqueFamousWitnesses returns the unique famous witnesses
// for a particular round.
//
// If one creator has multiple famous witnesses within the same round,
// all famous witnesses belonging to that creator are excluded (review docs)
func (g *Graph) UniqueFamousWitnesses(round uint64, witnesses map[uint64][]EventID, fame map[EventID]FameResult) []EventID {
	var famous []EventID
	creatorCount := make(map[NodeID]int)

	// First pass:
	// find every famous witness and count
	// how many famous witnesses each creator has.
	for _, witness := range witnesses[round] {
		result, ok := fame[witness]

		if !ok || result.Fame != FameFamous {
			continue
		}

		event, ok := g.Get(witness)
		if !ok {
			continue
		}

		famous = append(famous, witness)

		creatorCount[event.Creator]++
	}

	// Second Pass:
	// retain only creators with exactly
	// one famous witness in this round.

	var unique []EventID

	for _, witness := range famous {
		event, ok := g.Get(witness)
		if !ok {
			continue
		}

		if creatorCount[event.Creator] != 1 {
			continue
		}

		unique = append(unique, witness)
	}

	return unique
}

func (g *Graph) RoundReceived(head EventID, x EventID, membership *Membership) (uint64, bool) {
	if membership == nil || membership.Len() == 0 {
		return 0, false
	}

	if !g.Has(head) || !g.Has(x) {
		return 0, false
	}

	// x must actually be in the history we're evaluating
	if !g.IsAncestor(x, head) {
		return 0, false
	}

	witnesses := g.WitnessesByRound(head, membership)
	fame := g.FameDecisions(head, membership)

	maxRound := maxWitnessRound(witnesses)

	for round := uint64(1); round <= maxRound; round++ {
		if !AllFameDecidedThrough(round, witnesses, fame) {
			break
		}

		judges := g.UniqueFamousWitnesses(round, witnesses, fame)

		// Avoid treating an empty judgement set as recieving every event
		if len(judges) == 0 {
			continue
		}

		received := true

		for _, judge := range judges {
			if !g.IsAncestor(x, judge) {
				received = false
				break
			}
		}

		if received {
			return round, true
		}
	}

	return 0, false
}

func maxWitnessRound(witnesses map[uint64][]EventID) uint64 {
	var max uint64

	for round := range witnesses {
		if round > max {
			max = round
		}
	}

	return max
}

func (g *Graph) FirstReceivedEvent(x, judge EventID) (Event, bool) {
	if !g.Has(x) || !g.Has(judge) {
		return Event{}, false
	}

	if !g.IsAncestor(x, judge) {
		return Event{}, false
	}

	currentID := judge

	for {
		current, ok := g.Get(currentID)
		if !ok {
			return Event{}, false
		}

		if current.SelfParent == nil {
			return current, true
		}

		if !g.IsAncestor(x, *current.SelfParent) {
			return current, true
		}

		currentID = *current.SelfParent
	}
}

func (g *Graph) ConsensusTimestamp(head EventID, x EventID, membership *Membership) (int64, bool) {
	roundReceived, ok := g.RoundReceived(head, x, membership)
	if !ok {
		return 0, false
	}

	witnesses := g.WitnessesByRound(head, membership)

	fame := g.FameDecisions(head, membership)
	judges := g.UniqueFamousWitnesses(roundReceived, witnesses, fame)

	if len(judges) == 0 {
		return 0, false
	}

	timestamps := make([]int64, 0, len(judges))

	for _, judge := range judges {
		firstReceived, ok := g.FirstReceivedEvent(x, judge)

		if !ok {
			return 0, false
		}

		timestamps = append(timestamps, firstReceived.Timestamp)
	}

	sort.Slice(
		timestamps,
		func(i, j int) bool {
			return timestamps[i] < timestamps[j]
		},
	)

	return timestamps[len(timestamps)/2], true
}

func (g *Graph) WhitenedSignature(
	event Event,
	roundReceived uint64,
	witnesses map[uint64][]EventID,
	fame map[EventID]FameResult,
) ([]byte, bool) {
	judges := g.UniqueFamousWitnesses(roundReceived, witnesses, fame)
	if len(judges) == 0 {
		return nil, false
	}

	whitened := make([]byte, len(event.Signature))
	copy(whitened, event.Signature)

	for _, judgeID := range judges {
		judge, ok := g.Get(judgeID)

		if !ok {
			return nil, false
		}

		if len(judge.Signature) != len(whitened) {
			return nil, false
		}

		for i := range whitened {
			whitened[i] ^= judge.Signature[i]
		}
	}

	return whitened, true
}

func (g *Graph) ConsensusOrder(head EventID, membership *Membership) []ConsensusEvent {
	if membership == nil || membership.Len() == 0 || !g.Has(head) {
		return nil
	}

	witnesses := g.WitnessesByRound(head, membership)
	fame := g.FameDecisions(head, membership)

	var ordered []ConsensusEvent

	for _, event := range g.History(head) {
		roundReceived, ok := g.RoundReceived(head, event.ID, membership)
		if !ok {
			continue
		}

		timestamp, ok := g.ConsensusTimestamp(head, event.ID, membership)
		if !ok {
			continue
		}

		whitened, ok := g.WhitenedSignature(event, roundReceived, witnesses, fame)
		if !ok {
			continue
		}

		ordered = append(ordered, ConsensusEvent{
			Event:              event,
			RoundReceived:      roundReceived,
			ConsensusTimestamp: timestamp,
			WhitenedSignature:  whitened,
		})
	}

	sort.Slice(
		ordered,
		func(i, j int) bool {
			left := ordered[i]
			right := ordered[j]

			if left.RoundReceived != right.RoundReceived {
				return left.RoundReceived < right.RoundReceived
			}

			if left.ConsensusTimestamp != right.ConsensusTimestamp {
				return left.ConsensusTimestamp < right.ConsensusTimestamp
			}

			return bytes.Compare(left.WhitenedSignature, right.WhitenedSignature) < 0
		},
	)

	return ordered
}
