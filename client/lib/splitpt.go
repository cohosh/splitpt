/*
Package splitpt_client implements the functionality necessary for a client to establish a connection with
a server using splitpt.
*/
package splitpt_client

import (
	tt "anticensorshiptrafficsplitting/splitpt/common/turbotunnel"
	"log"
	"net"
	"time"

	"github.com/txthinking/socks5"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
)

const (
	HOST = "217.162.72.192"
	PORT = "8888"
	TYPE = "tcp"
)

type Transport struct {
	config *ClientConfig
}

type ClientConfig struct {
	// NumPaths is how many different paths traffic will be split across
	// [AHL] TBD if this is a field that will stay but for now is a placeholder
	NumPaths int
}

func NewSplitPTClient(config *ClientConfig) (Transport, error) {

	return Transport{config: config}, nil

}

func (t *Transport) Dial() (*smux.Stream, error) {
	var cleanup []func()
	defer func() {
		for i := len(cleanup) - 1; i >= 0; i-- {
			cleanup[i]()
		}
	}()

	log.Printf("Starting new session")

	ptclient, err := ConnectToPT()
	if err != nil {
		log.Printf("Error connecting to pt")
		return nil, err
	}
	tcpaddr, err := net.ResolveTCPAddr("tcp", "localhost:9090")
	if err != nil {
		log.Printf("Error resolving tcp addr")
		return nil, err
	}
	ptconn, err := ptclient.DialWithLocalAddr("tcp", "", "localhost:9090", tcpaddr)
	if err != nil {
		log.Printf("Error dialing")
		return nil, err
	}

	// TurboTunnel
	sessionID := tt.NewSessionID()
	pconn := tt.NewRedialPacketConn(sessionID, ptconn)
	conn, err := kcp.NewConn2(pconn.RemoteAddr(), nil, 0, 0, pconn)
	if err != nil {
		return nil, err
	}
	log.Printf("SessionID: %v", sessionID)

	smuxConfig := smux.DefaultConfig()
	smuxConfig.Version = 2
	smuxConfig.KeepAliveTimeout = 1 * time.Minute
	smuxConfig.MaxReceiveBuffer = 4 * 1024 * 1024 // default is 4 * 1024 * 1024
	smuxConfig.MaxStreamBuffer = 1 * 1024 * 1024  // default is 65536
	sess, err := smux.Client(conn, smuxConfig)
	if err != nil {
		return nil, err
	}

	stream, err := sess.OpenStream()
	if err != nil {
		return nil, err
	}

	return stream, nil
	// AHL simple tcp server for now
	//tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	//if err != nil {
	//	os.Exit(1)
	//}
	//return net.DialTCP(TYPE, nil, tcpServer)

}

func ConnectToPT() (*socks5.Client, error) {
	// Make a connection to a client PT process
	// https://spec.torproject.org/pt-spec/per-connection-args.html
	client, err := socks5.NewClient("127.0.0.1:64538", "cert=xmK64YEbi2h1aZC5P5s7MyiUN8gmypIRDnaiRKmB4/qT0lGkaAglYlzKPrkpc4I2PHhVNg;iat-mode=0", "\x00", 60, 0)
	if err != nil {
		log.Printf("Error connecting to pt")
		return nil, err
	}

	return client, nil
}

/*
type SplitPTConn struct {
	*smux.Stream
}

func (conn *SplitPTConn) Close() error {}
*/
