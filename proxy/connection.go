package proxy

import (
	"fmt"
	"net"
	"time"
)

type ConnectionInfo struct {
	From, To             string
	BytesIn, BytesOut    uint64
	Created, LastUpdated time.Time
	proxy                *Proxy
}

func (ci *ConnectionInfo) Identity() string {
	return fmt.Sprintf("%s:%s", ci.From, ci.To)
}

func (ci *ConnectionInfo) SwapServerConnection(conn *net.TCPConn) {
	ci.proxy.SwapServerConnection(conn)
}

type InternalConnectionInfo struct {
	from, to                      string
	bytesIn, bytesOut             uint64
	bytesOutUpdate, bytesInUpdate chan uint64
	readChannel                   chan *ConnectionInfoRequest
	closeChannel                  chan bool
	created, lastUpdated          time.Time
	proxy                         *Proxy
}

type ConnectionInfoRequest struct {
	ResponseChannel chan *ConnectionInfo
}

const (
	UpdateBufferSize = 5
)

func NewConnectionInfo(from string, to string, proxy *Proxy) *InternalConnectionInfo {
	now := time.Now().UTC()

	info := &InternalConnectionInfo{
		from:           from,
		to:             to,
		bytesIn:        0,
		bytesOut:       0,
		bytesInUpdate:  make(chan uint64, UpdateBufferSize),
		bytesOutUpdate: make(chan uint64, UpdateBufferSize),
		readChannel:    make(chan *ConnectionInfoRequest),
		closeChannel:   make(chan bool),
		created:        now,
		lastUpdated:    now,
		proxy:          proxy,
	}
	go info.loop()
	return info
}

func (ci *InternalConnectionInfo) Identity() string {
	return fmt.Sprintf("%s:%s", ci.from, ci.to)
}

func (ci *InternalConnectionInfo) UpdateBytesIn(bytes uint64) {
	ci.bytesInUpdate <- bytes
}

func (ci *InternalConnectionInfo) UpdateBytesOut(bytes uint64) {
	ci.bytesOutUpdate <- bytes
}

func (ci *InternalConnectionInfo) Get() *ConnectionInfo {

	request := &ConnectionInfoRequest{
		ResponseChannel: make(chan *ConnectionInfo),
	}
	ci.readChannel <- request
	return <-request.ResponseChannel
}

func (ci *InternalConnectionInfo) Close() {
	ci.closeChannel <- true
}

func (ci *InternalConnectionInfo) loop() {
	for {
		select {
		case in := <-ci.bytesInUpdate:
			ci.bytesIn += in
			ci.lastUpdated = time.Now().UTC()
		case out := <-ci.bytesOutUpdate:
			ci.bytesOut += out
			ci.lastUpdated = time.Now().UTC()
		case read := <-ci.readChannel:
			read.ResponseChannel <- &ConnectionInfo{
				From:        ci.from,
				To:          ci.to,
				BytesIn:     ci.bytesIn,
				BytesOut:    ci.bytesOut,
				Created:     ci.created,
				LastUpdated: ci.lastUpdated,
				proxy:       ci.proxy,
			}
		case <-ci.closeChannel:
			close(ci.readChannel)
			close(ci.bytesInUpdate)
			close(ci.bytesOutUpdate)
			close(ci.closeChannel)
			return
		}
	}
}

type ConnectionInfoCollection []*ConnectionInfo
type ConnectionInfoPredicate func(*ConnectionInfo) bool

func (c ConnectionInfoCollection) Select(fn ConnectionInfoPredicate) ConnectionInfoCollection {
	var p ConnectionInfoCollection
	for _, v := range c {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}

type ByIdentity ConnectionInfoCollection

func (a ByIdentity) Len() int           { return len(a) }
func (a ByIdentity) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIdentity) Less(i, j int) bool { return a[i].Identity() < a[j].Identity() }
