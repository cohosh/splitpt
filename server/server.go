package main

import (
	//	"io"
	//	"io/ioutil"
	"log"
	"net"

	//	"net/url"
	"os"
	//	"os/signal"
	//	"strings"
	"flag"

	//	"sync"
	//	"syscall"
	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

// func proxy(local *net.TCPConn, conn net.Conn) {}

func handler(conn net.Conn, ptInfo pt.ServerInfo) error {
	defer conn.Close()
	or, err := pt.DialOr(&ptInfo, conn.RemoteAddr().String(), "splitpt")
	if err != nil {
		return err
	}
	defer or.Close()
	// [AHL] do something with or and conn
	return nil
}

func acceptLoop(ln net.Listener, ptInfo pt.ServerInfo) error {
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				continue
			}
			log.Printf("accept error: %s", err.Error())
			return err
		}
		go handler(conn, ptInfo)
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
			go acceptLoop(ln, ptInfo)
			pt.Smethod(bindaddr.MethodName, ln.Addr())
		default:
			pt.SmethodError(bindaddr.MethodName, "no such method")
			log.Printf("No such method")
		}
	}
	pt.SmethodsDone()
}
