package hashgraph

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidID          = errors.New("invalid event ID")
	ErrMissingSelfParent  = errors.New("missing self-parent")
	ErrMissingOtherParent = errors.New("missing other-parent")
	ErrInvalidSelfParent  = errors.New("invalid self-parent")
	ErrInvalidOtherParent = errors.New("invalid other-parent")
	ErrInvalidIndex       = errors.New("invalid event index")
	ErrInvalidParentShape = errors.New("invalid parent shape")
	ErrInvalidSignature   = errors.New("invalid event signature")
)

type Graph struct {
	events map[EventID]Event
}

func NewGraph() *Graph {
	return &Graph{
		events: make(map[EventID]Event),
	}
}

func (g *Graph) Add(event Event) error {
	if g.Has(event.ID) {
		return nil
	}

	if !event.ValidID() {
		return ErrInvalidID
	}

	if !event.ValidSignature() {
		return ErrInvalidSignature
	}

	// Genesis events have no parents
	if event.Index == 0 {
		if event.SelfParent != nil || event.OtherParent != nil {
			return ErrInvalidParentShape
		}

		g.events[event.ID] = event
		return nil
	}

	// Every non-genesis event must have both parents
	if event.SelfParent == nil || event.OtherParent == nil {
		return ErrInvalidParentShape
	}

	selfParent, ok := g.Get(*event.SelfParent)
	if !ok {
		return fmt.Errorf(
			"%w: %s",
			ErrMissingSelfParent,
			event.SelfParent.Short(),
		)
	}

	otherParent, ok := g.Get(*event.OtherParent)
	if !ok {
		return fmt.Errorf(
			"%w: %s",
			ErrMissingOtherParent,
			event.OtherParent.Short(),
		)
	}

	if selfParent.Creator != event.Creator {
		return ErrInvalidSelfParent
	}

	if event.Index != selfParent.Index+1 {
		return ErrInvalidIndex
	}

	if otherParent.Creator == event.Creator {
		return ErrInvalidOtherParent
	}

	g.events[event.ID] = event

	return nil
}

func (g *Graph) Get(id EventID) (Event, bool) {
	event, ok := g.events[id]
	return event, ok
}

func (g *Graph) Has(id EventID) bool {
	_, ok := g.events[id]
	return ok
}

func (g *Graph) Len() int {
	return len(g.events)
}
