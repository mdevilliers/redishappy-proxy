// +build tcpserver

package main

import (
	"flag"
	"net"
	"os"

	"github.com/mdevilliers/redishappy/services/logger"
)

var address string
var connectionCount = 0

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
		connectionCount++
		logger.Info.Printf("CurrentConnections %d", connectionCount)
		go handleRequest(conn)
	}

	var ch chan bool
	<-ch // blocks forever
}

func handleRequest(conn net.Conn) {

	buf := make([]byte, 4048)
	for {
		n, err := conn.Read(buf)
		logger.Info.Print(string(buf[:n]))

		if err != nil {
			logger.Info.Printf("Error reading %s", err.Error())
			connectionCount--
			logger.Info.Printf("CurrentConnections %d", connectionCount)
			return
		}
		conn.Write(buf[:n])
	}
}
