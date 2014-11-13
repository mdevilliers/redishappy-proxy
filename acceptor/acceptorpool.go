package acceptor

import (
	"sync"
)

type AcceptorPool struct {
	sync.RWMutex
	m map[string]*Acceptor
}

func NewAcceptorPool() *AcceptorPool {
	return &AcceptorPool{
		m: make(map[string]*Acceptor),
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
		pool, err := NewAcceptor(localAddress, remoteAddress)

		if err != nil {
			return nil, err
		}

		p.m[name] = pool
		return pool, nil
	}
}
