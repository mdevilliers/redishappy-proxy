package acceptor

import (
	"net"
	"sync"

	"github.com/mdevilliers/redishappy-proxy/proxy"
	"github.com/mdevilliers/redishappy/services/logger"
)

type Acceptor struct {
	localAddress, remoteAddress *net.TCPAddr
	listener                    *net.TCPListener
	sync.RWMutex
	started             bool
	quit                chan bool
	acceptedConnections chan AcceptedConnection
	registry            *proxy.Registry
}

type AcceptedConnection struct {
	connection *net.TCPConn
	err        error
}

func NewAcceptor(localAddress, remoteAddress string, registry *proxy.Registry) (*Acceptor, error) {

	laddr, err := net.ResolveTCPAddr("tcp", localAddress)

	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveTCPAddr("tcp", remoteAddress)

	if err != nil {
		return nil, err
	}

	return &Acceptor{
		localAddress:        laddr,
		remoteAddress:       raddr,
		started:             false,
		quit:                make(chan bool),
		acceptedConnections: make(chan AcceptedConnection),
		registry:            registry,
	}, nil
}

func (a *Acceptor) IsRunning() bool {

	a.Lock()
	defer a.Unlock()

	return a.started
}

func (a *Acceptor) Stop() {

	a.Lock()
	defer a.Unlock()

	a.quit <- true
	a.started = false
}

func (a *Acceptor) Start() error {

	a.RLock()
	defer a.RUnlock()

	// idemnepotent start
	if a.started {
		return nil
	}

	listener, err := net.ListenTCP("tcp", a.localAddress)

	if err != nil {
		return err
	}
	a.listener = listener

	go a.acceptorLoop()
	go a.loop()

	a.started = true
	return nil
}

func (a *Acceptor) loop() {

	for {

		quit := false

		select {
		case accepted := <-a.acceptedConnections:

			if accepted.err != nil {
				logger.Error.Printf("Error on acceptor channel : %s", accepted.err)

			} else {
				p := proxy.NewProxy(accepted.connection, a.localAddress, a.remoteAddress, a.registry)
				go p.Start()
			}

		case <-a.quit:

			a.listener.Close()

			//flush the last message out
			//maybe a race?
			<-a.acceptedConnections
			quit = true
		}
		if quit {
			break
		}
	}
}

func (a *Acceptor) acceptorLoop() {

	for {
		conn, err := a.listener.AcceptTCP()
		a.acceptedConnections <- AcceptedConnection{connection: conn, err: err}
	}
}
