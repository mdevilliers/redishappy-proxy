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

func (p *AcceptorPool) RemoveExistingAcceptor(name string) {
	p.Lock()
	defer p.Unlock()
	delete(p.m, name)
}

func (p *AcceptorPool) NewOrDefaultAcceptor(name, localAddress, remoteAddress string) (*Acceptor, error) {
	p.Lock()
	defer p.Unlock()

	a, found := p.m[name]

	if found {
		err := a.UpdateRemoteAddress(remoteAddress)

		if err != nil {
			return nil, err
		}
		return a, nil
	}

	b, err := NewAcceptor(localAddress, remoteAddress, p.registry)

	if err != nil {
		return nil, err
	}

	p.m[name] = b
	return b, nil

}
