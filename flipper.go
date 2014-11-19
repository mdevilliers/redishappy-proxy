package main

import (
	"fmt"

	"github.com/mdevilliers/redishappy-proxy/acceptor"
	"github.com/mdevilliers/redishappy-proxy/proxy"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

type ProxyFlipper struct {
	pool     *acceptor.AcceptorPool
	registry *proxy.Registry
}

func NewProxyFlipper() *ProxyFlipper {

	registry := proxy.NewRegistry()

	return &ProxyFlipper{
		pool:     acceptor.NewAcceptorPool(registry),
		registry: registry,
	}
}

func (pf *ProxyFlipper) InitialiseRunningState(state *types.MasterDetailsCollection) {
	logger.Info.Printf("InitialiseRunningState called : %s", util.String(state.Items()))

	for _, md := range state.Items() {
		go pf.startAcceptorPool(md.Name, md.ExternalPort, md.Ip, md.Port)
	}
}

func (pf *ProxyFlipper) Orchestrate(switchEvent types.MasterSwitchedEvent) {
	logger.Info.Printf("Orchestrate called : %s", util.String(switchEvent))

	outgoingConnection := fmt.Sprintf("%s:%d", switchEvent.OldMasterIp, switchEvent.OldMasterPort)

	filter := func(ci *proxy.ConnectionInfo) bool {
		return ci.To == outgoingConnection || ci.From == outgoingConnection
	}

	// TODO : close existing acceptor with name

	// TODO : spin up a new acceptor pool

	// TODO : swap over existing connections
	// this will get all open connections either from to to the old server
	// there might not be many
	pf.registry.GetConnectionsWithFilter(filter)

}

func (pf *ProxyFlipper) startAcceptorPool(name string, externalport int, host string, port int) {

	localAddress := fmt.Sprintf("%s:%d", "localhost", externalport)
	remoteAddress := fmt.Sprintf("%s:%d", host, port)

	logger.Info.Printf("Proxying from %v to %v\n", localAddress, remoteAddress)

	acceptor, err := pf.pool.NewOrDefaultAcceptor(name, localAddress, remoteAddress)

	if err != nil {
		logger.Error.Printf("Error creating new acceptor for %s -> %s", localAddress, remoteAddress)
	} else {
		go acceptor.Start()
	}

}
