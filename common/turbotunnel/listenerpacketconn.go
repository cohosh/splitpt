package turbotunnel

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

	"www.bamsoftware.com/git/turbotunnel-paper.git/example/turbotunnel/turbotunnel"
)

type ListenerPacketConn struct {
	ln net.Listener
	*turbotunnel.QueuePacketConn
}

func NewListenerPacketConn(ln net.Listener) *ListenerPacketConn {
	c := &ListenerPacketConn{
		ln,
		turbotunnel.NewQueuePacketConn(ln.Addr(), 1*time.Minute),
	}
	go func() {
		err := c.acceptConnections()
		if err != nil {
			log.Printf("acceptConnections: %v", err)
		}
	}()
	return c
}

func (c *ListenerPacketConn) acceptConnections() error {
	for {
		conn, err := c.ln.Accept()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {
				continue
			}
			return err
		}
		go func() {
			defer conn.Close()
			err := c.handleConnection(conn)
			if err != nil {
				log.Printf("handleConnection: %v", err)
			}
		}()
	}
}

func (c *ListenerPacketConn) handleConnection(conn net.Conn) error {
	// First read the client's session identifier.
	var sessionID turbotunnel.SessionID
	_, err := io.ReadFull(conn, sessionID[:])
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	done := make(chan struct{})
	go func() {
		defer wg.Done()
		defer close(done) // Signal the write loop to finish.
		for {
			p, err := turbotunnel.ReadPacket(conn)
			if err != nil {
				return
			}
			c.QueuePacketConn.QueueIncoming(p, sessionID)
		}
	}()
	go func() {
		defer wg.Done()
		defer conn.Close() // Signal the read loop to finish.
		for {
			select {
			case <-done:
				return
			case p, ok := <-c.QueuePacketConn.OutgoingQueue(sessionID):
				if ok {
					err := turbotunnel.WritePacket(conn, p)
					if err != nil {
						return
					}
				}
			}
		}
	}()

	wg.Wait()
	return nil
}

func (c *ListenerPacketConn) Close() error {
	err := c.ln.Close()
	err2 := c.QueuePacketConn.Close()
	if err == nil {
		err = err2
	}
	return err
}
