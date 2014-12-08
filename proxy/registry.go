package proxy

import (
	"sort"
	"sync"
)

type Registry struct {
	sync.RWMutex
	m          map[string]*InternalConnectionInfo
	statistics *RegistryStatistics
}

type RegistryStatistics struct {
	TotalNumberOfConnections   uint
	CurrentNumberOfConnections uint
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

func (r *Registry) RegisterConnection(from, to string) *InternalConnectionInfo {

	r.Lock()
	defer r.Unlock()
	info := NewConnectionInfo(from, to)

	r.m[info.Identity()] = info

	r.statistics.TotalNumberOfConnections++

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
		CurrentNumberOfConnections: uint(len(r.m)),
	}
}
