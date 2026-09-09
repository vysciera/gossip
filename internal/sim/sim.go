package sim

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"

	"gossip/internal/hashgraph"
	"gossip/internal/network"
)

type Interaction struct {
	Step uint64

	From *network.Node
	To   *network.Node
}

type Simulation struct {
	Nodes []*network.Node

	Membership *hashgraph.Membership

	rng   *rand.Rand
	clock int64
	step  uint64
}

func New(seed int64, names ...string) *Simulation {
	rng := rand.New(rand.NewSource(seed))
	nodes := make([]*network.Node, 0, len(names))

	ids := make([]hashgraph.NodeID, 0, len(names))

	var clock int64 = 1

	for i, name := range names {
		privateKey := deterministicKey(seed, uint64(i))
		node := network.NewNodeWithPrivateKey(name, privateKey, clock)

		clock++

		nodes = append(nodes, node)
		ids = append(ids, node.ID)
	}

	return &Simulation{
		Nodes:      nodes,
		Membership: hashgraph.NewMembership(ids...),
		rng:        rng,
		clock:      clock,
	}
}

func (s *Simulation) Step() (Interaction, error) {
	if len(s.Nodes) < 2 {
		return Interaction{}, nil
	}

	fromIndex := s.rng.Intn(len(s.Nodes))

	toIndex := s.rng.Intn(len(s.Nodes) - 1)
	if toIndex >= fromIndex {
		toIndex++
	}

	from := s.Nodes[fromIndex]
	to := s.Nodes[toIndex]

	s.clock++
	s.step++

	if err := network.GossipAt(
		from,
		to,
		s.clock,
	); err != nil {
		return Interaction{}, err
	}

	return Interaction{
		Step: s.step,
		From: from,
		To:   to,
	}, nil
}

func (s *Simulation) Run(steps int) error {
	for i := 0; i < steps; i++ {
		if _, err := s.Step(); err != nil {
			return err
		}
	}

	return nil
}

func deterministicKey(seed int64, index uint64) ed25519.PrivateKey {
	var input [16]byte

	binary.BigEndian.PutUint64(
		input[0:8],
		uint64(seed),
	)

	binary.BigEndian.PutUint64(
		input[8:16],
		index,
	)

	digest := sha256.Sum256(input[:])

	return ed25519.NewKeyFromSeed(digest[:])
}

func (s *Simulation) ConsensusPrefixesAgree() (int, bool) {
	if len(s.Nodes) == 0 {
		return 0, true
	}

	orders := make([][]hashgraph.ConsensusEvent, len(s.Nodes))

	minLength := -1

	for i, node := range s.Nodes {
		orders[i] = node.Graph.ConsensusOrder(
			node.Head,
			s.Membership,
		)

		if minLength == -1 || len(orders[i]) < minLength {
			minLength = len(orders[i])
		}
	}

	if minLength <= 0 {
		return 0, true
	}

	for position := 0; position < minLength; position++ {
		expected := orders[0][position].Event.ID

		for nodeIndex := 1; nodeIndex < len(orders); nodeIndex++ {
			actual := orders[nodeIndex][position].Event.ID

			if actual != expected {
				return position, false
			}
		}
	}

	return minLength, true
}
