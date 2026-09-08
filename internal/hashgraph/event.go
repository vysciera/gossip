package hashgraph

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type EventID [sha256.Size]byte

type Event struct {
	ID	EventID

	Creator	string
	Index	uint64

	SelfParent	*EventID
	OtherParent	*EventID
}

func NewEvent(creator string, index uint64, selfParent *EventID, otherParent *EventID) Event {
	event := Event{
		Creator:	creator,
		Index:	index,
		SelfParent:	cloneID(selfParent),
		OtherParent:	cloneID(otherParent),
	}

	event.ID = event.calculateID()

	return event
}

func (e Event) calculateID() EventID {
	var buf bytes.Buffer

	writeString(&buf, e.Creator)

	_ = binary.Write(
		&buf,
		binary.BigEndian,
		e.Index,
	)

	writeParent(&buf, e.SelfParent)
	writeParent(&buf, e.OtherParent)

	return sha256.Sum256(buf.Bytes())
}

func (e Event) ValidID() bool {
	return e.ID == e.calculateID()
}

func (id EventID) Short() string {
	return fmt.Sprintf("%x", id[:4])
}

func writeString(buf *bytes.Buffer, value string) {
	_ = binary.Write(
		buf,
		binary.BigEndian,
		uint64(len(value)),
	)

	buf.WriteString(value)
}

func writeParent(buf *bytes.Buffer, parent *EventID) {
	if parent == nil {
		buf.WriteByte(0)
		return
	}

	buf.WriteByte(1)
	buf.Write(parent[:])
}

func cloneID(id *EventID) *EventID {
	if id == nil {
		return nil
	}

	copy := *id
	return &copy
}

