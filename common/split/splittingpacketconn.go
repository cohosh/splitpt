package split

import (
	"errors"
	"net"
	"time"
)

var errClosed = errors.New("operation on closed connection")
var errNotImplemented = errors.New("not implemented")

// stringAddr satisfies the net.Addr interface using fixed strings for the
// Network and String methods.
type stringAddr struct{ network, address string }

func (addr stringAddr) Network() string { return addr.network }
func (addr stringAddr) String() string  { return addr.address }

// Implements net.PacketConn interface with additional functionality for SplittingPacketConn

type SplittingPacketConn interface {
	ReadFrom(p []byte) (n int, addr net.Addr, err error)
	WriteTo(p []byte, addr net.Addr) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	getConn() net.Conn
	loop() error
	exchange() error
}
