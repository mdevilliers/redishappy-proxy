// +build proxy

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
var right1 string
var right2 string

func init() {
	flag.StringVar(&left, "left", "127.0.0.1:9999", "Incoming connection")
	flag.StringVar(&right1, "right1", "127.0.0.1:9998", "Outgoing connection 1")
	flag.StringVar(&right2, "right2", "127.0.0.1:9997", "Outgoing connection 2")
}

func main() {

	flag.Parse()

	registry := proxy.NewRegistry()

	acceptorPool := acceptor.NewAcceptorPool(registry)
	acceptor, _ := acceptorPool.NewOrDefaultAcceptor("default", left, right1)
	acceptor.Start()

	go func() {
		for {

			time.Sleep(time.Second)

			connections := registry.GetConnections()
			logger.Info.Printf("Connections : %s", util.StringPrettify(connections))
		}

	}()

	go func() {
		i := 1
		for {

			time.Sleep(5 * time.Second)

			var next string
			if i%2 == 0 {
				next = right1
			} else {
				next = right2
			}

			logger.Info.Printf("Swapping new connection to %s", next)
			acceptor.UpdateRemoteAddress(next)
			i++
		}

	}()

	var ch chan bool
	<-ch // blocks forever
}
