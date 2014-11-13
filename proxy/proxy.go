package proxy

import (
	"fmt"
	"log"
	"net"

	"github.com/mdevilliers/redishappy/services/logger"
)

type Proxy struct {
	identity     string
	lconn, rconn *net.TCPConn
	laddr, raddr *net.TCPAddr

	sentBytes     uint64
	receivedBytes uint64

	closeChannel chan bool
}

func NewProxy(conn *net.TCPConn, laddr *net.TCPAddr, raddr *net.TCPAddr) *Proxy {
	ident := fmt.Sprintf("%s:%s", laddr.String(), raddr.String())
	return &Proxy{
		identity:     ident,
		lconn:        conn,
		laddr:        laddr,
		raddr:        raddr,
		closeChannel: make(chan bool),
	}
}

func (p *Proxy) Start() {
	defer p.lconn.Close()

	//connect to remote
	// TODO : get this cnnectionf from a connection pool
	rconn, err := net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		logger.Info.Printf("Remote connection failed: %s", err)
		return
	}

	p.rconn = rconn
	defer p.rconn.Close()

	logger.Info.Printf("%s : Open", p.identity)

	//bidirectional copy
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)

	//wait for close...
	<-p.closeChannel

	// log.Printf("Closed (%d bytes sent, %d bytes recieved)", p.sentBytes, p.receivedBytes)
	logger.Info.Printf("%s : Closed", p.identity)
}

func (p *Proxy) pipe(src, dst *net.TCPConn) {

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
			logger.Info.Printf("%s : Read failed %s", p.identity, err)
			p.closeChannel <- true
			return
		}
		b := buff[:n]

		log.Printf(f, n)

		//write out result
		n, err = dst.Write(b)
		if err != nil {
			logger.Info.Printf("%s : Write failed %s", p.identity, err)
			p.closeChannel <- true
			return
		}
		// this needs a lock!
		// if islocal {
		// 	p.sentBytes += uint64(n)
		// } else {
		// 	p.receivedBytes += uint64(n)
		// }
	}
}
