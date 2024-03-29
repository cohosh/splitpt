// Client transport plugin for splitpt
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"

	//	"path/filepath"
	//	"strconv"
	//	"strings"
	"sync"
	"syscall"

	spt "anticensorshiptrafficsplitting/splitpt/client/lib"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
)

// Exchanges bytes between SOCKS connection and splitpt connection
// TODO [AHL] This will eventuall have to copy packets to different proxies according
// to the splitting algorithm being used
func copyLoop(socks, sptconn io.ReadWriter) {
	done := make(chan struct{}, 2)
	go func() {
		if _, err := io.Copy(socks, sptconn); err != nil {
			log.Printf("copying to SOCKS resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
	go func() {
		if _, err := io.Copy(sptconn, socks); err != nil {
			log.Printf("copying to SOCKS resulted in error: %v", err)
			done <- struct{}{}
		}
	}()
	<-done
	log.Printf("copy loop done")
}

func socksAcceptLoop(ln *pt.SocksListener, config spt.ClientConfig, shutdown chan struct{}, wg *sync.WaitGroup) error {
	log.Printf("socksAcceptLoop()")
	defer ln.Close()
	for {
		conn, err := ln.AcceptSocks()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				pt.Log(pt.LogSeverityError, "accept error: "+err.Error())
				continue
			}
		}
		log.Printf("SOCKS accepted %v", conn.Req)
		wg.Add(1)

		go func() {
			defer wg.Done()
			transport, err := spt.NewSplitPTClient(&config)
			if err != nil {
				log.Printf("Transport error: %s", err)
				conn.Reject()
				return
			}

			sconn, err := transport.Dial()
			if err != nil {
				log.Printf("Dial error: %s", err)
				conn.Reject()
				return
			}

			conn.Grant(nil)
			defer sconn.Close()
			copyLoop(conn, sconn)
		}()
	}
	return nil
}

func handler(conn *pt.SocksConn) error {
	log.Printf("handler()")
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

	var logOutput = ioutil.Discard
	if *logFilename != "" {
		logFile, err := os.OpenFile(*logFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		defer logFile.Close()
		logOutput = logFile
	}
	log.SetOutput(logOutput)
	log.Println("--- Starting SplitPT ---")
	// splitpt setup

	config := spt.ClientConfig{
		NumPaths: 2,
	}

	// begin goptlib client process
	ptInfo, err := pt.ClientSetup(nil)
	if err != nil {
		log.Printf("ClientSetup failed")
		os.Exit(1)
	}

	if ptInfo.ProxyURL != nil {
		pt.ProxyError(fmt.Sprintf("proxy %s is not supported", ptInfo.ProxyURL))
		log.Printf("Proxy is nor supported")
		os.Exit(1)
	}

	listeners := make([]net.Listener, 0)
	shutdown := make(chan struct{})
	var wg sync.WaitGroup

	for _, methodName := range ptInfo.MethodNames {
		switch methodName {
		case "splitpt":
			log.Printf("splitpt method found")
			ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
			if err != nil {
				pt.CmethodError(methodName, err.Error())
				break
			}
			log.Printf("Started SOCKS listenener at %v", ln.Addr())
			go socksAcceptLoop(ln, config, shutdown, &wg)
			pt.Cmethod(methodName, ln.Version(), ln.Addr())
			listeners = append(listeners, ln)
		default:
			log.Printf("no such method splitpt")
			pt.CmethodError(methodName, "no such method")
		}
	}
	pt.CmethodsDone()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	if os.Getenv("TOR_PT_EXIT_ON_STIN_CLOSE") == "1" {
		// This environment variable means we should treat EOF on stdin
		// just like SIGTERM: https://bugs.torproject.org/15435
		go func() {
			if _, err := io.Copy(ioutil.Discard, os.Stdin); err != nil {
				log.Printf("calling io.Copy(ioutil.Discard, osStdin) returned error: %v", err)
			}
			log.Printf("synthesizing SIGTERM because of stdin close")
			sigChan <- syscall.SIGTERM
		}()
	}

	// Wait for a signal.
	<-sigChan
	log.Printf("stopping splitpt")

	for _, ln := range listeners {
		ln.Close()
	}
	close(shutdown)
	wg.Wait()
	log.Println("SplitPT is done")

}
