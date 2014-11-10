package main

import (

	//	"fmt"
	"log"
	"net"
)

type proxy struct {
	lconn, rconn *net.TCPConn
	laddr, raddr *net.TCPAddr

	sentBytes     uint64
	receivedBytes uint64

	// errd         bool
	closeChannel chan bool
}

func (p *proxy) start() {
	defer p.lconn.Close()
	//connect to remote
	rconn, err := net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		log.Printf("Remote connection failed: %s", err)
		return
	}

	p.rconn = rconn
	defer p.rconn.Close()

	log.Printf("Opened %s >>> %s", p.lconn.RemoteAddr().String(), p.rconn.RemoteAddr().String())

	//bidirectional copy
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)

	//wait for close...
	<-p.closeChannel

	// log.Printf("Closed (%d bytes sent, %d bytes recieved)", p.sentBytes, p.receivedBytes)
	log.Print("Channel Closed")
}

func (p *proxy) pipe(src, dst *net.TCPConn) {

	var f string
	islocal := src == p.lconn

	if islocal {
		f = ">>> %d bytes sent"
	} else {
		f = "<<< %d bytes recieved"
	}

	//directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			log.Printf("Read failed '%s'\n", err)
			p.closeChannel <- true
			return
		}
		b := buff[:n]

		log.Printf(f, n)

		//write out result
		n, err = dst.Write(b)
		if err != nil {
			log.Printf("Write failed '%s'\n", err)
			p.closeChannel <- true
			return
		}
		// if islocal {
		// 	p.sentBytes += uint64(n)
		// } else {
		// 	p.receivedBytes += uint64(n)
		// }
	}
}
