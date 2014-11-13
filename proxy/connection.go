package proxy

import (
	"fmt"
	"time"
)

type ConnectionInfo struct {
	From, To             string
	BytesIn, BytesOut    uint64
	Created, LastUpdated time.Time
}

type InternalConnectionInfo struct {
	from, to                      string
	bytesIn, bytesOut             uint64
	bytesOutUpdate, bytesInUpdate chan uint64
	readChannel                   chan *ConnectionInfoRequest
	created, lastUpdated          time.Time
}

type ConnectionInfoRequest struct {
	ResponseChannel chan *ConnectionInfo
}

const (
	UpdateBufferSize = 5
)

func NewConnectionInfo(from string, to string) *InternalConnectionInfo {
	now := time.Now().UTC()

	info := &InternalConnectionInfo{
		from:           from,
		to:             to,
		bytesIn:        0,
		bytesOut:       0,
		bytesInUpdate:  make(chan uint64, UpdateBufferSize),
		bytesOutUpdate: make(chan uint64, UpdateBufferSize),
		readChannel:    make(chan *ConnectionInfoRequest),
		created:        now,
		lastUpdated:    now,
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
			}
		}
	}
}
