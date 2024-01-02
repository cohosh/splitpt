// Client transport plugin for splitpt
package main

import (
	"flag"
	"fmt"
	"io"
//	"io/ioutil"
	"log"
	"net"
	"os"
//	"os/signal"
//	"path/filepath"
//	"strconv"
//	"strings"
//	"sync"
//	"syscall"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	spt "anticensorshiptrafficsplitting/splitpt/client/lib"
)

// Exchanges bytes between SOCKS connection and splitpt connection
// [AHL] unused fn bc i haven't setup event handling
func copyLoop(socks, sptconn io.ReadWriter) {
	done := make(chan struct{}, 2)
	go func() {
		if _, err := io.Copy(socks, sptconn); err != nil {
			log.Printf("copying to SOCKS resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
	go func()  {
		if _, err := io.Copy(sptconn, socks); err != nil {
			log.Printf("copying to SOCKS resulted in error: %v", err)
		done <- struct{}{}
		}
	}()
	<- done
	log.Printf("copy loop done")
}

// TODO handle the socks cxn between splitpt and the pt being used to transport traffic
func socksAcceptLoop(ln *pt.SocksListener, config spt.ClientConfig) {
	defer ln.Close()
	for {
		conn, err := ln.AcceptSocks()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				pt.Log(pt.LogSeverityError, "accept error: " + err.Error())
				continue
			}
		}
		log.Printf("SOCKS accepted %v", conn.Req)
		// [AHL] should be replaced with copyLoop later
		go handler(conn)
	}	
}

func handler(conn *pt.SocksConn) error {
	defer conn.Close()
	remote, err := net.Dial("tcp", conn.Req.Target)
	if err != nil {
		conn.Reject()
		log.Printf("Dialing error: %v", err)
		return err
	}
	defer remote.Close()
	err = conn.Grant(remote.RemoteAddr().(*net.TCPAddr))
	if err != nil {
		log.Printf("Connection error: %v", err)
		return err
	}
	// [AHL] do something with conn and remote
	return nil
}

func main() {
	logFilename := flag.String("log", "", "name of log file")	
	flag.Parse()


	// Logging
	log.SetFlags(log.LstdFlags | log.LUTC)

	if *logFilename != "" {
		logFile, err := os.OpenFile(*logFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		defer logFile.Close()
	}

	// splitpt setup

	config := spt.ClientConfig {
		NumPaths: 2,
	}
	
	// begin goptlib client process
	ptInfo, err := pt.ClientSetup(nil)
	if err != nil {
		os.Exit(1)
	}

	if ptInfo.ProxyURL != nil {
		pt.ProxyError(fmt.Sprintf("proxy %s is not supported", ptInfo.ProxyURL))
		os.Exit(1)
	}
	for _, methodName := range ptInfo.MethodNames { 
		// [AHL] I think methodnames is fetched from the torrc?
		switch methodName {
			case "splitpt":
				ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
				if err != nil {
					pt.CmethodError(methodName, err.Error())
					break
				}
				go socksAcceptLoop(ln, config)
				pt.Cmethod(methodName, ln.Version(), ln.Addr())
			default:
				pt.CmethodError(methodName, "no such method")
		}
	}
	pt.CmethodsDone()

	log.Println("SplitPT is done")

}
