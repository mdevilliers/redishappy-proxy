package proxy

import (
	"fmt"
	"net"
	"sync"

	"github.com/mdevilliers/redishappy/services/logger"
)

type Proxy struct {
	sync.RWMutex
	connectionInfo                      *InternalConnectionInfo
	leftCloseChannel, rightCloseChannel chan bool
	registry                            *Registry
	left, right                         *Pipe
	ownedConnection                     *net.TCPConn
	fromConnection                      *net.TCPConn
	started                             bool
}

func NewProxy(conn *net.TCPConn, conn2 *net.TCPConn, registry *Registry) *Proxy {

	leftCloseChannel := make(chan bool)
	rightCloseChannel := make(chan bool)

	from := conn.RemoteAddr().String()
	to := conn2.RemoteAddr().String()

	proxy := &Proxy{
		rightCloseChannel: rightCloseChannel,
		leftCloseChannel:  leftCloseChannel,
		registry:          registry,
		ownedConnection:   conn2,
		fromConnection:    conn,
		started:           false,
	}

	proxy.left = NewPipe(conn, proxy.ownedConnection, DirectionLeftToRight, leftCloseChannel, proxy)
	proxy.right = NewPipe(proxy.ownedConnection, conn, DirectionRightToLeft, rightCloseChannel, proxy)
	proxy.connectionInfo = registry.RegisterConnection(from, to, proxy)

	return proxy
}

func (p *Proxy) Start() {

	p.Lock()

	if p.started {
		message := fmt.Sprintf("Proxy already started :%s", p.identity())
		logger.Error.Printf(message)
		panic(message)
	}

	go p.left.Open()
	go p.right.Open()

	p.started = true
	p.Unlock()

	logger.Info.Printf("%s : Open", p.identity())

	select {
	case <-p.leftCloseChannel:
		p.right.Close()

	case <-p.rightCloseChannel:
		p.left.Close()
	}

	p.ownedConnection.Close()

	p.registry.UnRegisterConnection(p.identity())
	logger.Info.Printf("%s : Closed", p.identity())
}

func (p *Proxy) SwapServerConnection(conn *net.TCPConn) {

	p.Lock()
	defer p.Unlock()

	p.left.SwapServerConnection(conn)
	p.right.SwapServerConnection(conn)

	p.ownedConnection.Close()
	p.ownedConnection = conn

	p.registry.UpdateExistingConnection(p.identity(), conn.RemoteAddr().String())
}

func (p *Proxy) UpdateStatistics(direction Direction, amount uint64) {

	if direction == DirectionLeftToRight {
		p.connectionInfo.UpdateBytesOut(amount)
	} else {
		p.connectionInfo.UpdateBytesIn(amount)
	}
}

func (p *Proxy) identity() string {
	return p.connectionInfo.Identity()
}
