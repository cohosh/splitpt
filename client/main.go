package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"sync"
	"syscall"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
)

func main() {
    ptInfo, err :=  pt.ClientSetup(nil)
    if err != nil {
	log.Fatal(err)
    }
    if ptInfo.ProxyURL != nil {
	pt.ProxyError("proxy is not supported")
	os.Exit(1)
    }

    listeners := make([]net.Listener, 0)
    shutdown := make(chan struct{})
    var wg sync.WaitGroup

    for _, methodName := range ptInfo.MethodNames {
        switch methodName {
            case "splitpt":
                ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
                if err != nil {
	        	pt.CmethodError(methodName, err.Error())
			break
		}
                log.Printf("Started splitpt SOCKS listener at %v.", ln.Addr())
        }
    }

}
