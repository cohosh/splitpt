package turbotunnel

import (
	"encoding/binary"
	"io"
)

// ReadPacket decapsulates a packet from r. It returns io.EOF if and only if
// there were zero bytes to be read from r.
func ReadPacket(r io.Reader) ([]byte, error) {
	var length uint16
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	p := make([]byte, length)
	_, err = io.ReadFull(r, p)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return p, err
}

// WritePacket encapsulates a packet into w. It panics if the length of the p
// cannot be represented by a uint16.
func WritePacket(w io.Writer, p []byte) error {
	length := uint16(len(p))
	if int(length) != len(p) {
		panic("packet too long")
	}
	err := binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return err
	}
	_, err = w.Write(p)
	return err
}
