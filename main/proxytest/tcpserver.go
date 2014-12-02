// +build tcpserver

package main

import (
	"flag"
	"net"
	"os"

	"github.com/mdevilliers/redishappy/services/logger"
)

var address string

func init() {
	flag.StringVar(&address, "server", "127.0.0.1:9998", "Server address")
}

func main() {

	flag.Parse()

	l, err := net.Listen("tcp", address)
	if err != nil {
		logger.Info.Printf("Error listening:", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	logger.Info.Printf("Listening on %s", address)
	for {

		conn, err := l.Accept()
		if err != nil {
			logger.Info.Printf("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleRequest(conn)
	}

	var ch chan bool
	<-ch // blocks forever
}

func handleRequest(conn net.Conn) {

	for {

		buf := make([]byte, 1024)

		_, err := conn.Read(buf)

		if err != nil {
			logger.Info.Printf("Error reading %s", err.Error())
			return
		}

		conn.Write([]byte("Message received."))

	}
}
