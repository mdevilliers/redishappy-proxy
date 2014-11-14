package main

import (
	"flag"

	"github.com/mdevilliers/redishappy-proxy/acceptor"
	"github.com/mdevilliers/redishappy-proxy/proxy"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/util"

	"time"
)

var left string
var right string

func init() {
	flag.StringVar(&left, "left", "127.0.0.1:9999", "Incoming connection")
	flag.StringVar(&right, "right", "127.0.0.1:6379", "Outgoing connection")
}

func main() {

	flag.Parse()

	registry := proxy.NewRegistry()

	acceptorPool := acceptor.NewAcceptorPool(registry)
	acceptor, _ := acceptorPool.NewOrDefaultAcceptor("default", left, right)
	acceptor.Start()

	go func() {
		for {

			time.Sleep(time.Second)

			connections := registry.GetConnections()
			logger.Info.Printf("Connections : %s", util.String(connections))
		}

	}()

	var ch chan bool
	<-ch // blocks forever
}
