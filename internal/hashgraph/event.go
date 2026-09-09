package hashgraph

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type EventID [sha256.Size]byte
type NodeID [ed25519.PublicKeySize]byte

type Event struct {
	ID EventID

	Creator NodeID
	Index   uint64

	Timestamp int64

	SelfParent  *EventID
	OtherParent *EventID

	Signature []byte
}

func NewEvent(privateKey ed25519.PrivateKey, index uint64, timestamp int64, selfParent, otherParent *EventID) Event {
	publicKey := privateKey.Public().(ed25519.PublicKey)

	var creator NodeID
	copy(creator[:], publicKey)

	event := Event{
		Creator:     creator,
		Index:       index,
		Timestamp:   timestamp,
		SelfParent:  cloneID(selfParent),
		OtherParent: cloneID(otherParent),
	}

	event.ID = event.calculateID()

	event.Signature = ed25519.Sign(
		privateKey,
		event.ID[:],
	)

	return event
}

func (e Event) calculateID() EventID {
	var buf bytes.Buffer
	buf.Write(e.Creator[:])

	_ = binary.Write(
		&buf,
		binary.BigEndian,
		e.Index,
	)

	_ = binary.Write(
		&buf,
		binary.BigEndian,
		e.Timestamp,
	)

	writeParent(&buf, e.SelfParent)
	writeParent(&buf, e.OtherParent)

	return sha256.Sum256(buf.Bytes())
}

func (e Event) ValidID() bool {
	return e.ID == e.calculateID()
}

func (e Event) ValidSignature() bool {
	return ed25519.Verify(
		ed25519.PublicKey(e.Creator[:]),
		e.ID[:],
		e.Signature,
	)
}

func (id EventID) Short() string {
	return fmt.Sprintf("%x", id[:4])
}

func (id NodeID) Short() string {
	return fmt.Sprintf("%x", id[:4])
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
