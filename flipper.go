package main

import (
	"fmt"
	"net"

	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

type ProxyFlipper struct {
}

func NewProxyFlipper() *ProxyFlipper {
	return &ProxyFlipper{}
}

func (pf *ProxyFlipper) InitialiseRunningState(state *types.MasterDetailsCollection) {
	logger.Info.Printf("InitialiseRunningState called : %s", util.String(state.Items()))

	for _, md := range state.Items() {
		go pf.startAcceptorPool(md.Name, md.ExternalPort, md.Ip, md.Port)
	}
}

func (*ProxyFlipper) Orchestrate(switchEvent types.MasterSwitchedEvent) {
	logger.Info.Printf("Orchestrate called : %s", util.String(switchEvent))
}

func (pf *ProxyFlipper) startAcceptorPool(name string, externalport int, host string, port int) {

	localAddr := fmt.Sprintf("%s:%d", "localhost", externalport)
	remoteAddr := fmt.Sprintf("%s:%d", host, port)

	logger.Info.Printf("Proxying from %v to %v\n", localAddr, remoteAddr)

	laddr, err := net.ResolveTCPAddr("tcp", localAddr)
	panicIfError(err)
	raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	panicIfError(err)
	listener, err := net.ListenTCP("tcp", laddr)
	panicIfError(err)

	// acceptor pool
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logger.Error.Printf("Failed to accept connection '%s'\n", err)
			continue
		}

		p := NewProxy(conn, laddr, raddr)

		go p.start()
	}
}

func panicIfError(err error) {
	if err != nil {
		logger.Error.Panic(err.Error())
	}
}
