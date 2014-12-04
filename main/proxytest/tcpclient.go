// +build tcpclient

package main

import (
	"flag"
	"net"
	"os"
	"time"

	"github.com/mdevilliers/redishappy/services/logger"
)

var server string

func init() {
	flag.StringVar(&server, "server", "127.0.0.1:9999", "Server address")
}

func main() {

	flag.Parse()

	str := "Hello World"

	tcpAddr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		logger.Info.Printf("ResolveTCPAddr failed: %s", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logger.Info.Printf("DialTCP failed: %s", err.Error())
		os.Exit(1)
	}

	go func() {

		for {
			_, err = conn.Write([]byte(str))
			if err != nil {
				logger.Info.Printf("Write to server failed:  %s", err.Error())
				os.Exit(1)
			}

			reply := make([]byte, 1024)

			_, err = conn.Read(reply)
			if err != nil {
				logger.Info.Printf("Read from server failed: %s", err.Error())
				os.Exit(1)
			}

			//logger.Info.Print(numberOfBytes)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	var ch chan bool
	<-ch // blocks forever
}
