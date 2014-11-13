package proxy

import (
	"sync"
)

type Registry struct {
	sync.RWMutex
	m map[string]*InternalConnectionInfo
}

func NewRegistry() *Registry {
	return &Registry{
		m: make(map[string]*InternalConnectionInfo),
	}
}

func (r *Registry) RegisterConnection(from string, to string) *InternalConnectionInfo {

	r.Lock()
	defer r.Unlock()
	info := NewConnectionInfo(from, to)

	r.m[info.Identity()] = info
	return info
}

func (r *Registry) UnRegisterConnection(identity string) {

	r.Lock()
	defer r.Unlock()

	delete(r.m, identity)
}
