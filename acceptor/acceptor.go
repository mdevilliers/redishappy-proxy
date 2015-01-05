package acceptor

import (
	"errors"
	"net"
	"sync"
	"time"

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

var StopRequested = errors.New("Listener stopped via Stop() method.")

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

func (a *Acceptor) UpdateRemoteAddress(remoteAddress string) error {
	a.Lock()
	defer a.Unlock()

	raddr, err := net.ResolveTCPAddr("tcp", remoteAddress)

	if err != nil {
		return err
	}

	a.remoteAddress = raddr
	return nil
}

func (a *Acceptor) loop() {

	for {

		select {
		case accepted := <-a.acceptedConnections:

			if accepted.err == StopRequested {
				break
			}

			if accepted.err != nil {
				logger.Error.Printf("Error on acceptor channel : %s", accepted.err)
			} else {

				// TODO : get this connection from a connection pool
				conn, err := net.DialTCP("tcp", nil, a.remoteAddress)
				if err != nil {
					logger.Error.Printf("Remote connection failed: %s", err)
				} else {
					p := proxy.NewProxy(accepted.connection, conn, a.registry)
					go p.Start()
				}

			}
		}
	}
}

func (a *Acceptor) acceptorLoop() {

	for {

		a.listener.SetDeadline(time.Now().Add(time.Second))

		conn, err := a.listener.AcceptTCP()

		select {
		case <-a.quit:
			a.acceptedConnections <- AcceptedConnection{connection: conn, err: StopRequested}
			a.listener.Close()
			break
		default:
			// do nothing as no request to stop made
		}

		if err != nil {
			netErr, ok := err.(net.Error)

			//If this is a timeout, then continue to wait for
			//new connections
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			}
		}

		a.acceptedConnections <- AcceptedConnection{connection: conn, err: err}
	}
}
