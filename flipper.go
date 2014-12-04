package main

import (
	"fmt"
	"net"

	"github.com/mdevilliers/redishappy-proxy/acceptor"
	"github.com/mdevilliers/redishappy-proxy/proxy"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

type ProxyFlipper struct {
	pool          *acceptor.AcceptorPool
	registry      *proxy.Registry
	configuration *configuration.ConfigurationManager
}

func NewProxyFlipper(config *configuration.ConfigurationManager) *ProxyFlipper {

	registry := proxy.NewRegistry()

	return &ProxyFlipper{
		pool:     acceptor.NewAcceptorPool(registry),
		registry: registry,
	}
}

func (pf *ProxyFlipper) InitialiseRunningState(state *types.MasterDetailsCollection) {
	logger.NoteWorthy.Printf("InitialiseRunningState : %s", util.String(state.Items()))

	for _, md := range state.Items() {
		go pf.ensureCorrectAcceptorPoolIsRunning(md.Name, md.ExternalPort, md.Ip, md.Port)
	}
}

func (pf *ProxyFlipper) Orchestrate(switchEvent types.MasterSwitchedEvent) {
	logger.NoteWorthy.Printf("Orchestrate : %s", util.String(switchEvent))

	config := pf.configuration.GetCurrentConfiguration()
	cluster, err := config.FindClusterByName(switchEvent.Name)

	if err != nil {
		logger.Info.Printf("Unknown cluster : %s, error : %s", switchEvent.Name, err.Error())
		return
	}

	// close existing acceptor with name
	// spin up a new acceptor pool
	pf.ensureCorrectAcceptorPoolIsRunning(switchEvent.Name, cluster.ExternalPort, switchEvent.NewMasterIp, switchEvent.NewMasterPort)

	// swap out existing connection endpoint
	// without dropping connections
	oldEndpoint := fmt.Sprintf("%s:%d", switchEvent.OldMasterIp, switchEvent.OldMasterPort)
	newEndpoint := fmt.Sprintf("%s:%d", switchEvent.NewMasterIp, switchEvent.NewMasterPort)

	filter := func(ci *proxy.ConnectionInfo) bool {
		return ci.To == oldEndpoint || ci.From == oldEndpoint
	}

	existingconnections := pf.registry.GetConnectionsWithFilter(filter)

	for _, connection := range existingconnections {

		laddr, _ := net.ResolveTCPAddr("tcp", newEndpoint)
		conn, err := net.DialTCP("tcp", nil, laddr)

		if err != nil {
			logger.Error.Printf("Remote connection failed: %s", err)
		}

		connection.SwapServerConnection(conn)
	}

}

func (pf *ProxyFlipper) ensureCorrectAcceptorPoolIsRunning(name string, externalport int, host string, port int) {

	localAddress := fmt.Sprintf("%s:%d", "localhost", externalport)
	remoteAddress := fmt.Sprintf("%s:%d", host, port)

	logger.Info.Printf("Proxying from %v to %v\n", localAddress, remoteAddress)

	acceptor, err := pf.pool.ReplaceOrDefaultAcceptor(name, localAddress, remoteAddress)

	if err != nil {
		logger.Error.Printf("Error creating new acceptor for %s -> %s", localAddress, remoteAddress)
	} else {
		go acceptor.Start()
	}
}
