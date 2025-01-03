package split

import (
	"bufio"
	"errors"
	"log"
	"math/rand/v2"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"anticensorshiptrafficsplitting/splitpt/common/turbotunnel"
	tt "anticensorshiptrafficsplitting/splitpt/common/turbotunnel"
)

var errClosed = errors.New("operation on closed connection")
var errNotImplemented = errors.New("not implemented")

// stringAddr satisfies the net.Addr interface using fixed strings for the
// Network and String methods.
type stringAddr struct{ network, address string }

func (addr stringAddr) Network() string { return addr.network }
func (addr stringAddr) String() string  { return addr.address }

// RandomPacketConn implements the net.PacketConn interface by continually
//
// Every Turbo Tunnel design will need some sort of PacketConn adapter that
// adapts the session layer's sequence of packets to the obfuscation layer. But
// not every such adapter will look like RoundRobinPacketConn. It depends on what
// the obfuscation layer looks like. Some obfuscation layers will not need a
// persistent connection. One could, for example, handle every ReadFrom or
// WriteTo as an independent network operation.
type RandomPacketConn struct {
	sessionID  turbotunnel.SessionID
	remoteAddr net.Addr
	recvQueue  chan []byte
	sendQueue  chan []byte
	closeOnce  sync.Once
	closed     chan struct{}
	connList   []net.Conn
	// What error to return when the RoundRobinPacketConn is closed.
	err atomic.Value
}

func NewRandomPacketConn(
	sessionID tt.SessionID,
	connList []net.Conn,
	remote net.Addr,
) *RoundRandomConn {
	c := &RoundRobinPacketConn{
		sessionID:  sessionID,
		remoteAddr: remote,
		recvQueue:  make(chan []byte, 32),
		sendQueue:  make(chan []byte, 32),
		closed:     make(chan struct{}),
	}
	go func() {
		c.closeWithError(c.loop())
	}()
	return c
}

// Next returns the next connection to write a packet to
func (c *RandomPacketConn) getConn() net.Conn {
	index := rand.IntN(len(c.connList))
	return c.connList[index]
}

// loop dials c.remoteAddr in a loop, exchanging packets on each new connection
// as long as it lasts. Only errors in dialing break the loop and report the
// error to the caller.
func (c *RandomPacketConn) loop() error {
	for {
		select {
		case <-c.closed:
			return nil
		default:
		}
		log.Printf("[RR Packet Conn] session %v: redialing %v", c.sessionID, c.remoteAddr)
		err := c.exchange()
		if err != nil {
			return err
		}
	}
}

func (c *RandomPacketConn) exchange() error {

	// Begin by sending the session identifier to each connection; everything after that is
	// encapsulated packets.
	for _, conn := range c.connList {
		_, err := conn.Write(c.sessionID[:])
		if err != nil {
			// TODO: Because we don't currently have a redial mechanism,
			// errors are fatal
			log.Printf("[RR Packet Conn] Error writing to conn %s", err.Error())
			return err
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	done := make(chan struct{})
	// Read encapsulated packets from the connection and write them to
	// c.recvQueue.
	for _, conn := range c.connList {
		go func() {
			defer wg.Done()
			defer close(done) // Signal the write loop to finish.
			br := bufio.NewReader(conn)
			for {
				p, err := turbotunnel.ReadPacket(br)
				if err != nil {
					return
				}
				select {
				case <-c.closed:
					return
				case c.recvQueue <- p:
				}
			}
		}()
	}
	// Read packets from c.sendQueue and encapsulate them into the
	// connection.
	go func() {
		defer wg.Done()
		for _, conn := range c.connList {
			defer conn.Close() // Signal the read loop to finish.
		}
		for {
			select {
			case <-c.closed:
				return
			case <-done:
				return
			case p := <-c.sendQueue:
				conn := c.getConn()
				bw := bufio.NewWriter(conn)
				err := turbotunnel.WritePacket(bw, p)
				if err != nil {
					return
				}
				err = bw.Flush()
				if err != nil {
					return
				}
			}
		}
	}()

	// Exchange packets until the connection is terminated.
	wg.Wait()
	return nil
}

func (c *RandomPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	select {
	case <-c.closed:
		return 0, nil, &net.OpError{Op: "read", Net: c.remoteAddr.Network(), Source: c.sessionID, Addr: c.remoteAddr, Err: c.err.Load().(error)}
	default:
	}
	select {
	case <-c.closed:
		return 0, nil, &net.OpError{Op: "read", Net: c.remoteAddr.Network(), Source: c.sessionID, Addr: c.remoteAddr, Err: c.err.Load().(error)}
	case buf := <-c.recvQueue:
		return copy(p, buf), c.remoteAddr, nil
	}
}

func (c *RandomPacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	select {
	case <-c.closed:
		return 0, &net.OpError{Op: "write", Net: c.remoteAddr.Network(), Source: c.sessionID, Addr: c.remoteAddr, Err: c.err.Load().(error)}
	default:
	}
	// Copy the slice so that the caller may reuse p.
	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case c.sendQueue <- buf:
	default: // Silently drop outgoing packets if the send queue is full.
	}
	return len(buf), nil
}

// closeWithError unblocks pending operations and makes future operations fail
// with the given error. If err is nil, it becomes errClosed.
func (c *RandomPacketConn) closeWithError(err error) error {
	firstClose := false
	c.closeOnce.Do(func() {
		firstClose = true
		// Store the error that will be returned for future operations.
		if err == nil {
			err = errClosed
		}
		c.err.Store(err)
		close(c.closed)
	})
	if !firstClose {
		return &net.OpError{Op: "close", Net: c.remoteAddr.Network(), Source: c.sessionID, Addr: c.remoteAddr, Err: c.err.Load().(error)}
	}
	return nil
}

func (c *RandomPacketConn) Close() error { return c.closeWithError(nil) }

func (c *RandomPacketConn) LocalAddr() net.Addr  { return c.sessionID }
func (c *RandomPacketConn) RemoteAddr() net.Addr { return c.remoteAddr }

func (c *RandomPacketConn) SetDeadline(t time.Time) error      { return errNotImplemented }
func (c *RandomPacketConn) SetReadDeadline(t time.Time) error  { return errNotImplemented }
func (c *RandomPacketConn) SetWriteDeadline(t time.Time) error { return errNotImplemented }
