package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var localAddr = flag.String("l", ":9999", "local address")
var remoteAddr = flag.String("r", "localhost:80", "remote address")

func main() {

	flag.Parse()
	fmt.Printf("Proxying from %v to %v\n", *localAddr, *remoteAddr)

	laddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	panicIfError(err)
	raddr, err := net.ResolveTCPAddr("tcp", *remoteAddr)
	panicIfError(err)
	listener, err := net.ListenTCP("tcp", laddr)
	panicIfError(err)

	// acceptor pool
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Printf("Failed to accept connection '%s'\n", err)
			continue
		}

		p := &proxy{
			lconn:        conn,
			laddr:        laddr,
			raddr:        raddr,
			closeChannel: make(chan bool),
		}
		go p.start()
	}
}

func panicIfError(err error) {
	if err != nil {
		log.Panic(err.Error())
	}
}
