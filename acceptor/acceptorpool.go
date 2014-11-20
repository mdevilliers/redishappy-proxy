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
	a, ok := p.m[name]
	if ok {
		return a, nil
	} else {
		a, err := NewAcceptor(localAddress, remoteAddress, p.registry)

		if err != nil {
			return nil, err
		}

		p.m[name] = a
		return a, nil
	}
}

func (p *AcceptorPool) RemoveExistingAcceptor(name string) {
	p.Lock()
	defer p.Unlock()
	delete(p.m, name)
}

func (p *AcceptorPool) ReplaceOrDefaultAcceptor(name, localAddress, remoteAddress string) (*Acceptor, error) {
	p.Lock()
	defer p.Unlock()

	a, found := p.m[name]

	if found {
		go a.Stop()
		delete(p.m, name)
	}

	b, err := NewAcceptor(localAddress, remoteAddress, p.registry)

	if err != nil {
		return nil, err
	}

	p.m[name] = b
	return b, nil

}
