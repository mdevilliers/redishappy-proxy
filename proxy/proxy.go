package proxy

import (
	"io"
	"net"

	"github.com/mdevilliers/redishappy/services/logger"
)

type Proxy struct {
	connectionInfo *InternalConnectionInfo
	lconn, rconn   *net.TCPConn
	laddr, raddr   *net.TCPAddr

	sentBytes     uint64
	receivedBytes uint64

	closeChannel chan bool
	registry     *Registry
}

func NewProxy(conn *net.TCPConn, laddr *net.TCPAddr, raddr *net.TCPAddr, registry *Registry) *Proxy {
	return &Proxy{
		connectionInfo: registry.RegisterConnection(conn.RemoteAddr().String(), raddr.String()),
		lconn:          conn,
		laddr:          laddr,
		raddr:          raddr,
		closeChannel:   make(chan bool),
		registry:       registry,
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

	logger.Info.Printf("%s : Open", p.connectionInfo.Identity())

	//bidirectional copy
	go p.pipe(true, p.lconn, p.rconn)
	go p.pipe(false, p.rconn, p.lconn)

	//wait for close...
	<-p.closeChannel

	p.registry.UnRegisterConnection(p.connectionInfo.Identity())
	logger.Info.Printf("%s : Closed", p.connectionInfo.Identity())
}

func (p *Proxy) pipe(islocal bool, src io.Reader, dst io.Writer) {

	//directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)

		if err != nil {
			logger.Info.Printf("%s : Read failed %s", p.connectionInfo.Identity(), err)
			p.closeChannel <- true
			return
		}

		b := buff[:n]
		n, err = dst.Write(b)

		if err != nil {
			logger.Info.Printf("%s : Write failed %s", p.connectionInfo.Identity(), err)
			p.closeChannel <- true
			return
		}

		if islocal {
			p.connectionInfo.UpdateBytesOut(uint64(n))
		} else {
			p.connectionInfo.UpdateBytesIn(uint64(n))
		}
	}
}
