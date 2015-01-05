package proxy

import (
	"sort"
	"sync"
	"sync/atomic"
)

type Registry struct {
	sync.RWMutex
	m          map[string]*InternalConnectionInfo
	statistics *RegistryStatistics
}

type RegistryStatistics struct {
	TotalNumberOfConnections   uint32
	CurrentNumberOfConnections uint32
}

func NewRegistry() *Registry {
	return &Registry{
		m: make(map[string]*InternalConnectionInfo),
		statistics: &RegistryStatistics{
			TotalNumberOfConnections:   0,
			CurrentNumberOfConnections: 0,
		},
	}
}

func (r *Registry) RegisterConnection(from, to string, proxy *Proxy) *InternalConnectionInfo {

	r.Lock()
	defer r.Unlock()
	info := NewConnectionInfo(from, to, proxy)

	r.m[info.Identity()] = info

	atomic.AddUint32(&r.statistics.TotalNumberOfConnections, 1)

	return info
}

func (r *Registry) UpdateExistingConnection(identity, to string) {

	r.Lock()
	defer r.Unlock()

	connectionInfo, ok := r.m[identity]

	if ok {
		connectionInfo.to = to
		newIdentity := connectionInfo.Identity()
		delete(r.m, identity)
		r.m[newIdentity] = connectionInfo
	}
}

func (r *Registry) UnRegisterConnection(identity string) {

	r.Lock()
	defer r.Unlock()

	connectionInfo, ok := r.m[identity]

	if ok {
		go connectionInfo.Close()
		delete(r.m, identity)
	}
}

func (r *Registry) GetConnections() ConnectionInfoCollection {

	r.RLock()
	defer r.RUnlock()

	arr := make([]*ConnectionInfo, 0, len(r.m))
	for _, value := range r.m {
		arr = append(arr, value.Get())
	}

	sort.Sort(ByIdentity(arr))
	return arr
}

func (r *Registry) GetConnectionsWithFilter(filter ConnectionInfoPredicate) ConnectionInfoCollection {
	return r.GetConnections().Select(filter)
}

func (r *Registry) GetStatistics() RegistryStatistics {
	r.Lock()
	defer r.Unlock()

	return RegistryStatistics{
		TotalNumberOfConnections:   r.statistics.TotalNumberOfConnections,
		CurrentNumberOfConnections: uint32(len(r.m)),
	}
}
