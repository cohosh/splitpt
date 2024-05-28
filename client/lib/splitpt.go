/*
Package splitpt_client implements the functionality necessary for a client to establish a connection with
a server using splitpt.
*/
package splitpt_client

import (
	tt "anticensorshiptrafficsplitting/splitpt/common/turbotunnel"
	"bufio"
	"context"
	"log"
	"net"
	"os/exec"
	"strings"
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

	ptchan := make(chan string)
	pterr := make(chan error)
	ptshutdown := make(chan struct{})

	ctx := context.Background()
	ptproc := exec.CommandContext(ctx, "lyrebird", "-enableLogging", "-logLevel", "DEBUG")
	//log.Printf(ptproc.Env)
	ptproc.Env = append(ptproc.Environ(), "TOR_PT_MANAGED_TRANSPORT_VER=1")
	ptproc.Env = append(ptproc.Environ(), "TOR_PT_CLIENT_TRANSPORTS=obfs4")
	ptproc.Env = append(ptproc.Environ(), "TOR_PT_STATE_LOCATION=../pt-setup/client-state/")
	ptprocout, err := ptproc.StdoutPipe()
	if err != nil {
		log.Printf("Error getting stdout pipe")
		pterr <- err
	}
	err1 := ptproc.Start()
	if err1 != nil {
		log.Printf("Error starting PT process")
		pterr <- err1
	}

	go func() {
		scanner := bufio.NewScanner(ptprocout)
		for scanner.Scan() {
			log.Printf("scanned: ")
			log.Printf(scanner.Text())
			if strings.Contains(scanner.Text(), "socks5") {
				line := strings.Split(scanner.Text(), " ")
				ptchan <- line[3]
				log.Printf("Got SOCKS5 addr")
				break
			} else {
				continue
			}
		}
		if err2 := scanner.Err(); err2 != nil {
			log.Printf("Error scanning for socks5 addr")
			pterr <- err2
		}
		<-ptshutdown
		err3 := ptproc.Wait()
		if err3 != nil {
			log.Printf("Error completing command: %s", err3.Error())
		}
		log.Printf("PT Process Exited")
	}()

	var socks5addr string
	select {
	case socks5addr = <-ptchan:
		//socks5addr := <-ptchan
		log.Printf("SOCKS5 addr: %s", socks5addr)
	case err := <-pterr:
		return nil, err
	}

	client, err := socks5.NewClient(socks5addr, "cert=xmK64YEbi2h1aZC5P5s7MyiUN8gmypIRDnaiRKmB4/qT0lGkaAglYlzKPrkpc4I2PHhVNg;iat-mode=0", "\x00", 60, 0)
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
