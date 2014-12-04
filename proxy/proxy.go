package proxy

import (
	"net"
	"sync"

	"github.com/mdevilliers/redishappy/services/logger"
)

type Proxy struct {
	connectionInfo                      *InternalConnectionInfo
	leftCloseChannel, rightCloseChannel chan bool
	registry                            *Registry
	left, right                         *Pipe
	ownedConnection                     *net.TCPConn
	fromConnection                      *net.TCPConn
	sync.RWMutex
}

func NewProxy(conn *net.TCPConn, conn2 *net.TCPConn, registry *Registry) *Proxy {

	leftCloseChannel := make(chan bool)
	rightCloseChannel := make(chan bool)

	from := conn.RemoteAddr().String()
	to := conn2.RemoteAddr().String()

	connectionInfo := registry.RegisterConnection(from, to)

	proxy := &Proxy{
		connectionInfo:    connectionInfo,
		rightCloseChannel: rightCloseChannel,
		leftCloseChannel:  leftCloseChannel,
		registry:          registry,
		ownedConnection:   conn2,
		fromConnection:    conn,
	}

	proxy.left = NewPipe(conn, proxy.ownedConnection, DirectionLeftToRight, leftCloseChannel, proxy)
	proxy.right = NewPipe(proxy.ownedConnection, conn, DirectionRightToLeft, rightCloseChannel, proxy)

	connectionInfo.RegisterProxy(proxy)

	return proxy
}

func (p *Proxy) Start() {

	go p.left.Open()
	go p.right.Open()

	logger.Info.Printf("%s : Open", p.Identity())

	select {
	case <-p.leftCloseChannel:
		p.right.Close()

	case <-p.rightCloseChannel:
		p.left.Close()
	}

	p.ownedConnection.Close()

	p.registry.UnRegisterConnection(p.Identity())
	logger.Info.Printf("%s : Closed", p.Identity())
}

func (p *Proxy) SwapServerConnection(conn *net.TCPConn) {

	p.Lock()
	defer p.Unlock()

	p.left.SwapServerConnection(conn)
	p.right.SwapServerConnection(conn)

	p.ownedConnection.Close()
	p.ownedConnection = conn

	p.registry.UnRegisterConnection(p.Identity())

	from := p.fromConnection.RemoteAddr().String()
	p.connectionInfo = p.registry.RegisterConnection(from, conn.RemoteAddr().String())
	p.connectionInfo.RegisterProxy(p)
}

func (p *Proxy) UpdateStatistics(direction Direction, amount uint64) {

	if direction == DirectionLeftToRight {
		p.connectionInfo.UpdateBytesOut(amount)
	} else {
		p.connectionInfo.UpdateBytesIn(amount)
	}
}

func (p *Proxy) Identity() string {
	return p.connectionInfo.Identity()
}
