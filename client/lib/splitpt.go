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
	"net"
//	"net/url"
//	"strings"
//	"time"
	"os"

//	"github.com/pion/ice/v2"
//	"github.com/xtaci/smux"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)
type Transport struct {}

type ClientConfig struct {
	// NumPaths is how many different paths traffic will be split across
	// [AHL] TBD if this is a field that will stay but for now is a placeholder
	NumPaths int
}

func NewSplitPTClient(config ClientConfig) (*Transport, error) {

	log.Println("\n\n\n --- Starting SplitPT Client ---")

	transport := &Transport{}
	return transport, nil

}

func (t *Transport) Dial() (net.Conn, error) {
	var cleanup []func()
	defer func() {
		for i := len(cleanup) - 1; i >= 0; i-- {
			cleanup[i]()
		}
	}()

	log.Printf("Starting new session")

	// AHL simple tcp server for now
	tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		os.Exit(1)
	}
	conn, err := net.DialTCP(TYPE, nil, tcpServer)
	return conn, nil


}
/*
type SplitPTConn struct {
	*smux.Stream
}

func (conn *SplitPTConn) Close() error {}
*/

