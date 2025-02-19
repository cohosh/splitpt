package splitpt_client

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/txthinking/socks5"
)

func LyrebirdConnect(path string, args []string, cert string) (*socks5.Client, error) {
	log.Printf("Conecting to Lyrebird")
	ptchan := make(chan string)
	pterr := make(chan error)
	ptshutdown := make(chan struct{})

	ctx := context.Background()
	ptproc := exec.CommandContext(ctx, path)
	//log.Printf(ptproc.Env)
	cwd, _ := os.Getwd()
	statedir := fmt.Sprintf("TOR_PT_STATE_LOCATION=%s/pt-setup/client-state/", cwd)
	ptproc.Env = append(ptproc.Environ(), "TOR_PT_MANAGED_TRANSPORT_VER=1")
	ptproc.Env = append(ptproc.Environ(), "TOR_PT_EXIT_ON_STDIN_CLOSE=0")
	ptproc.Env = append(ptproc.Environ(), "TOR_PT_CLIENT_TRANSPORTS=proteus")
	ptproc.Env = append(ptproc.Environ(), statedir)

	log.Printf("Getting stdoutpipe")
	ptprocout, err := ptproc.StdoutPipe()
	if err != nil {
		log.Printf("Error getting stdout pipe")
		pterr <- err
	}
	log.Printf("Starting ptproc")
	err1 := ptproc.Start()
	if err1 != nil {
		log.Printf("Error starting PT process: %s", err1.Error())
		pterr <- err1
	}

	log.Printf("Scanning for SOCKS address")
	go func() {
		scanner := bufio.NewScanner(ptprocout)
		for scanner.Scan() {
			log.Println(scanner.Text())
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
			log.Printf("Error scanning for socks5 addr: %s", err2.Error())
			pterr <- err2
		}
		go func() {
			for scanner.Scan() {
				log.Println(scanner.Text())
			}
		}()

		<-ptshutdown
		err3 := ptproc.Wait()
		if err3 != nil {
			log.Printf("Error completing command: %s", err3.Error())
		}
	}()

	var socks5addr string
	select {
	case socks5addr = <-ptchan:
		log.Printf("SOCKS5 addr: %s", socks5addr)
	case err := <-pterr:
		log.Printf("pterr had something in it")
		return nil, err
	}

	log.Printf("Getting Lyrebird SOCKS client")
	client, err := socks5.NewClient(socks5addr, encodeArgs(args), "\x00", 0, 0)
	if err != nil {
		log.Printf("Error connecting to pt")
		return nil, err
	}

	return client, nil
}

func encodeArgs(args []string) string {
	return strings.Join(args[:], ";")
}
