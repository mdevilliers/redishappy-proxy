package proxy

import (
	"bytes"
	"strings"
	"testing"
)

type MockStatisticGatherer struct {
	totalin  uint64
	totalout uint64
}

func (m *MockStatisticGatherer) UpdateStatistics(d Direction, num uint64) {
	if d == DirectionLeftToRight {
		m.totalout = m.totalout + num
	} else {
		m.totalin = m.totalin + num
	}
}

func TestBasicPipe(t *testing.T) {

	closeChannel := make(chan bool)
	gatherer := &MockStatisticGatherer{totalin: 0, totalout: 0}

	buffer := []byte{}
	buff := bytes.NewBuffer(buffer)

	str := "hello, good evening and nothing much!"
	left := strings.NewReader(str)

	pipe := NewPipe(left, buff, DirectionLeftToRight, closeChannel, gatherer)

	pipe.Close()
	pipe.Open()

	if buff.String() != str {
		t.Errorf("Buffer not copied - should be %s but is  %s", str, buff.String())
	}

	if int(gatherer.totalout) != len(str) {
		t.Errorf("Statistics in should be %d but are %d", len(str), gatherer.totalout)
	}

}
