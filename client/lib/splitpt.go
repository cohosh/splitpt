/*
Package splitpt_client implements the functionality necessary for a client to establish a connection with
a server using splitpt.
*/
package splitpt_client

import (
	//	"context"
	//	"errors"
	"log"
	//	"math/rand"
//	"net"
	//	"net/url"
	//	"strings"
		"time"
//	"os"
	//	"github.com/pion/ice/v2"
	"github.com/xtaci/smux"
	"github.com/xtaci/kcp-go/v5"
	tt "anticensorshiptrafficsplitting/splitpt/common/turbotunnel"


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
	
	// TurboTunnel
	dummyaddr := ""
	sessionID := tt.NewSessionID()
	pconn := tt.NewRedialPacketConn(sessionID, dummyaddr)
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
	defer sess.Close()

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

/*
type SplitPTConn struct {
	*smux.Stream
}

func (conn *SplitPTConn) Close() error {}
*/
