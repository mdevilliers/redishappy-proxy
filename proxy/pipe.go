package proxy

import "io"

type Direction int

const (
	DirectionLeftToRight = 1
	DirectionRightToLeft = 2
)

type Pipe struct {
	statisticGatherer StatisticGatherer
	left              io.Reader
	right             io.Writer
	direction         Direction
	closeChannel      chan bool
	safelyIgnoreError bool
	safelyClose       bool
}

type StatisticGatherer interface {
	UpdateStatistics(Direction, uint64)
}

func NewPipe(left io.Reader, right io.Writer, direction Direction, closeChannel chan bool, stat StatisticGatherer) *Pipe {
	return &Pipe{
		left:              left,
		right:             right,
		direction:         direction,
		statisticGatherer: stat,
		closeChannel:      closeChannel,
		safelyIgnoreError: false,
	}
}

func (p *Pipe) SwapServerConnection(conn io.ReadWriter) {

	p.safelyIgnoreError = true

	if p.direction == DirectionLeftToRight {
		p.right = conn
	} else {
		p.left = conn
	}

}

func (p *Pipe) Open() {

	buff := make([]byte, 0xffff)
	for {

		n, err := p.left.Read(buff)

		if err != nil {

			if !p.safelyIgnoreError {
				//logger.Info.Printf("%s : Read failed %d, %s", p.proxy.Identity(), p.direction, err)
				p.closeChannel <- true
				return
			}
		}

		b := buff[:n]
		n, err = p.right.Write(b)

		if err != nil {
			if !p.safelyIgnoreError {
				//logger.Info.Printf("%s : Write failed %s", p.proxy.Identity(), err)
				p.closeChannel <- true
				return
			}
		}

		p.statisticGatherer.UpdateStatistics(p.direction, uint64(n))
		p.safelyIgnoreError = false

		if p.safelyClose {
			return
		}
	}
}

func (p *Pipe) Close() {
	p.safelyClose = true
}
