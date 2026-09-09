package network

import (
	"crypto/ed25519"
	"crypto/rand"
	"time"

	"smalltalk/internal/hashgraph"
)

type Node struct {
	Name string

	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey

	ID hashgraph.NodeID

	Graph *hashgraph.Graph
	Head  hashgraph.EventID

	NextIndex uint64
}

func NewNode(name string) *Node {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	var id hashgraph.NodeID
	copy(id[:], publicKey)

	graph := hashgraph.NewGraph()
	genesis := hashgraph.NewEvent(
		privateKey,
		0,
		time.Now().UnixNano(),
		nil,
		nil,
	)

	if err := graph.Add(genesis); err != nil {
		panic(err)
	}

	return &Node{
		Name:       name,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		ID:         id,
		Graph:      graph,
		Head:       genesis.ID,
		NextIndex:  1,
	}
}
