package acceptor

import (
	"sync"

	"github.com/mdevilliers/redishappy-proxy/proxy"
)

type AcceptorPool struct {
	sync.RWMutex
	m        map[string]*Acceptor
	registry *proxy.Registry
}

func NewAcceptorPool(registry *proxy.Registry) *AcceptorPool {
	return &AcceptorPool{
		m:        make(map[string]*Acceptor),
		registry: registry,
	}
}

func (p *AcceptorPool) NewOrDefaultAcceptor(name, localAddress, remoteAddress string) (*Acceptor, error) {

	p.Lock()
	defer p.Unlock()

	// if exists return existing else create a new one
	pool, ok := p.m[name]
	if ok {
		return pool, nil
	} else {
		pool, err := NewAcceptor(localAddress, remoteAddress, p.registry)

		if err != nil {
			return nil, err
		}

		p.m[name] = pool
		return pool, nil
	}
}
