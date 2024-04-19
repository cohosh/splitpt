package main

import (
	"errors"
	"io"
	"log"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"flag"
	"os"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"

	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"

	tt "anticensorshiptrafficsplitting/splitpt/common/turbotunnel"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

func proxy(local *net.TCPConn, stream *smux.Stream) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if _, err := io.Copy(stream, local); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			log.Printf("error copying from ORPort %v", err)
		}
		local.CloseRead()
		stream.Close()
		wg.Done()
	}()
	go func() {
		if _, err := io.Copy(local, stream); err != nil && !errors.Is(err, io.EOF) {
			log.Printf("error copying to ORPort %v", err)
		}
		local.CloseWrite()
		stream.Close()
		wg.Done()
	}()

	wg.Wait()

}

func handler(stream *smux.Stream, ptInfo pt.ServerInfo) error {
	defer stream.Close()
	or, err := pt.DialOr(&ptInfo, stream.RemoteAddr().String(), "splitpt")
	if err != nil {
		return err
	}
	defer or.Close()
	proxy(or, stream)

	return nil
}

func acceptLoop(kcpln *kcp.Listener, ptInfo pt.ServerInfo) error {
	defer kcpln.Close()
	for {
		conn, err := kcpln.AcceptKCP()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				continue
			}
			log.Printf("accept error: %s", err.Error())
			return err
		}
		go func() {
			defer conn.Close()
			err := acceptStreams(conn, ptInfo)
			if err != nil {
				log.Printf("Error: %s", err)
			}
		}()
	}
	return nil
}

func acceptStreams(conn *kcp.UDPSession, ptInfo pt.ServerInfo) error {
	smuxConfig := smux.DefaultConfig()
	smuxConfig.Version = 2
	smuxConfig.KeepAliveTimeout = 1 * time.Minute
	sess, err := smux.Server(conn, smuxConfig)
	if err != nil {
		return err
	}
	defer sess.Close()

	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {
				continue
			}
			return err
		}

		go func() {
			defer stream.Close()
			err := handler(stream, ptInfo)
			if err != nil {
				log.Printf("Error: %s", err)
			}
		}()
	}
	return nil

}

func main() {
	// Setup logging
	logFileName := flag.String("log", "", "log file to write to")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.LUTC)

	if *logFileName != "" {
		f, err := os.OpenFile(*logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("can't open log file: %s", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Printf("Starting")

	ptInfo, err := pt.ServerSetup(nil)
	if err != nil {
		log.Printf("Error setting up server: %s", err)
		os.Exit(1)
	}

	for _, bindaddr := range ptInfo.Bindaddrs {
		switch bindaddr.MethodName {
		case "splitpt":
			ln, err := net.ListenTCP("tcp", bindaddr.Addr)
			if err != nil {
				log.Printf("Error: %s", err.Error())
				break
			}

			// TurboTunnel
			pconn := tt.NewListenerPacketConn(ln)
			kcpln, err := kcp.ServeConn(nil, 0, 0, pconn)
			if err != nil {
				log.Printf("Error: %s", err.Error())
				break
			}

			go acceptLoop(kcpln, ptInfo)
			pt.Smethod(bindaddr.MethodName, ln.Addr())

		default:
			pt.SmethodError(bindaddr.MethodName, "no such method")
			log.Printf("No such method")
		}
	}
	pt.SmethodsDone()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	<-sigChan
}
