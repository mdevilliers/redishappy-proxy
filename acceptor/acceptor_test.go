package acceptor

import (
	"testing"
	"time"
)

func TestWillNotAcceptInvalidHostPortCombos(t *testing.T) {

	_, err := NewAcceptor("rubbish", "1.1.1.1:8080")

	if err == nil {
		t.Error("Should not accept invalid addresses")
	}

	_, err = NewAcceptor("1.1.1.1:8080", "rubbish")

	if err == nil {
		t.Error("Should not accept invalid addresses")
	}
}

func TestAcceptorCanStartAndStop(t *testing.T) {

	acceptor, err := NewAcceptor("localhost:9089", "localhost:9090")

	if err != nil {
		t.Error("Acceptor should not error")
	}

	go acceptor.Start()

	time.Sleep(time.Millisecond * 10)

	if !acceptor.IsRunning() {
		t.Error("Acceptor pool should be running")
	}

	acceptor.Stop()

	time.Sleep(time.Millisecond * 10)

	if acceptor.IsRunning() {
		t.Error("Acceptor pool should have stopped")
	}

}

func TestUnableToStartTwoAcceptorsOnTheSameAddress(t *testing.T) {

	acceptor1, _ := NewAcceptor("localhost:9089", "localhost:9090")
	acceptor2, _ := NewAcceptor("localhost:9089", "localhost:9090")

	go func() {
		err := acceptor1.Start()

		if err != nil {
			t.Error("Acceptor should not error")
		}
	}()

	go func() {
		err := acceptor2.Start()

		if err == nil {
			t.Error("Acceptor should error")
		}
	}()

	time.Sleep(time.Millisecond * 10)
	acceptor1.Stop()
	time.Sleep(time.Millisecond * 10)

}

func TestUnableToStartAcceptorsTwice(t *testing.T) {

	acceptor1, _ := NewAcceptor("localhost:9089", "localhost:9090")
	err := acceptor1.Start()

	if err != nil {
		t.Errorf("Acceptor should not error : %s", err.Error())
	}

	time.Sleep(time.Millisecond * 10)
	err = acceptor1.Start()

	if err != nil {
		t.Error("Acceptor should not error")
	}

	time.Sleep(time.Millisecond * 10)
	acceptor1.Stop()
	time.Sleep(time.Millisecond * 10)
}
