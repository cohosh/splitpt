/*
Package splitpt_client implements the functionality necessary for a client to establish a connection with
a server using splitpt.
*/
package splitpt_client

import (
	split "anticensorshiptrafficsplitting/splitpt/common/split"
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

type SplitPTClient struct {
	config SplitPTClientConfig
}

func NewSplitPTClient(sptClientConfig split.SplitPTClientConfig) (SplitPTClient, error) {
	return SplitPTClient{config: config}, nil
}

func (t *SplitPTClient) Dial() (*smux.Stream, error) {
	var cleanup []func()
	defer func() {
		for i := len(cleanup) - 1; i >= 0; i-- {
			cleanup[i]()
		}
	}()

	log.Printf("Starting new session")

	connList, err := t.GetPTConnections()
	if err != nil {
		log.Printf("Error connecting to pts: %s", err.Error())
		return nil, err
	}

	log.Printf("Setting up turbotunnel")
	// TurboTunnel
	sessionID := tt.NewSessionID()
	// TODO make this the correct type of redial packat conn for the splitting algorithm
	/*var pconn
	switch sptConfig.SplittingAlg {
		case "round-robin":
			pconn = split.NewRoundRobinPacketConn(sessionID, ptconn)
	}*/
	pconn := split.NewRoundRobinPacketConn(sessionID, connList)
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

func (t *SplitPTClient) GetPTConnections() ([]net.Conn, error) {
	var connlist []net.Conn
	tcpaddr, err := net.ResolveTCPAddr("tcp", "localhost:9090")
	if err != nil {
		log.Printf("Error resolving TCP address: %s", err.Error())
		return nil, error
	}
	for conn := range t.config.Connections {
		var client *socks5.Client
		if conn.Transport == "lyrebird" {
			client := spt.LyrebirdConnect(conn.Args, conn.Cert)
		} else {
			err := error.New("Unrecognized PT")
			return _, err
		}
		ptconn, err := client.DialWithLocalAddr("tcp", "", "localhost:9090", tcpaddr)
		if err != nil {
			log.Printf("Error dialing: %s", err.Error())
			return nil, err
		}
		connList.append(ptconn)
	}
}
