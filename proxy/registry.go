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

func (r *Registry) GetConnections() []*ConnectionInfo {

	r.RLock()
	defer r.RUnlock()

	arr := make([]*ConnectionInfo, 0, len(r.m))
	for _, value := range r.m {
		arr = append(arr, value.Get())
	}
	return arr
}
