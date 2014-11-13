package acceptor

import (
	"github.com/mdevilliers/redishappy-proxy/proxy"
	"github.com/mdevilliers/redishappy/services/logger"
	"net"
	"sync"
)

type Acceptor struct {
	localAddress, remoteAddress *net.TCPAddr
	listener                    *net.TCPListener
	sync.RWMutex
	started             bool
	quit                chan bool
	acceptedConnections chan AcceptedConnection
}

type AcceptedConnection struct {
	connection *net.TCPConn
	err        error
}

func NewAcceptor(localAddress, remoteAddress string) (*Acceptor, error) {

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
				p := proxy.NewProxy(accepted.connection, a.localAddress, a.remoteAddress)
				go p.Start()
			}

		case <-a.quit:

			a.listener.Close()

			//flush the last message out
			//maybe a race?
			<-a.acceptedConnections
			quit = true

			logger.Info.Print("Stopped!")
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
