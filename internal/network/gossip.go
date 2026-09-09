package network

import (
	"errors"
	"time"

	"smalltalk/internal/hashgraph"
)

var ErrUnknownGossipHead = errors.New("unknown gossip head")

func Gossip(from, to *Node) error {
	return GossipHead(from, to, from.Head)
}

func GossipAt(from, to *Node, timestamp int64) error {
	return GossipHeadAt(
		from,
		to,
		from.Head,
		timestamp,
	)
}

func GossipHead(from *Node, to *Node, head hashgraph.EventID) error {
	return GossipHeadAt(
		from,
		to,
		head,
		time.Now().UnixNano(),
	)
}

func GossipHeadAt(from, to *Node, head hashgraph.EventID, timestamp int64) error {
	if !from.Graph.Has(head) {
		return ErrUnknownGossipHead
	}

	if err := copyMissingEvents(
		from.Graph,
		to.Graph,
		head,
	); err != nil {
		return err
	}

	selfParent := to.Head
	otherParent := head

	event := hashgraph.NewEvent(
		to.PrivateKey,
		to.NextIndex,
		timestamp,
		&selfParent,
		&otherParent,
	)

	if err := to.Graph.Add(event); err != nil {
		return err
	}

	to.Head = event.ID
	to.NextIndex++

	return nil
}

func copyMissingEvents(from *hashgraph.Graph, to *hashgraph.Graph, head hashgraph.EventID) error {
	for _, event := range from.History(head) {
		if to.Has(event.ID) {
			continue
		}

		if err := to.Add(event); err != nil {
			return err
		}
	}

	return nil
}
