package turbotunnel

import (
	"crypto/rand"
	"encoding/hex"
)

type SessionID [8]byte

func NewSessionID() SessionID {
	var id SessionID
	_, err := rand.Read(id[:])
	if err != nil {
		panic(err)
		// TODO tap in to PT logging here
	}

	return id
}

func (id SessionID) Network() string { return "session" }
func (id SessionID) String() string  { return hex.EncodeToString(id[:]) }
