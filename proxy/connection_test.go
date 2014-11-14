package proxy

import (
	"testing"
	"time"
)

func TestBasicUsage(t *testing.T) {
	connection := NewConnectionInfo("A", "B")

	if connection.Identity() != "A:B" {
		t.Errorf("identity should be A:B not %s", connection.Identity())
	}

	update := func() {
		for i := 1; i <= 10; i++ {
			go connection.UpdateBytesIn(100)
			go connection.UpdateBytesOut(100)
		}
	}
	go update()
	go update()
	go update()

	time.Sleep(time.Millisecond * 10)
	details := connection.Get()

	if details.From != "A" {
		t.Errorf("From should be 'A' not %s", details.From)
	}

	if details.To != "B" {
		t.Errorf("From should be 'B' not %s", details.To)
	}

	if details.BytesIn != 3000 {
		t.Errorf("Bytes in should by 3000 not %d", details.BytesIn)
	}
	if details.BytesOut != 3000 {
		t.Errorf("Bytes out should by 3000 not %d", details.BytesOut)
	}
}
