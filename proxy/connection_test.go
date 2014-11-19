package proxy

import (
	"sort"
	"sync"
	"testing"
)

func TestBasicUsage(t *testing.T) {
	connection := NewConnectionInfo("A", "B")

	if connection.Identity() != "A:B" {
		t.Errorf("identity should be A:B not %s", connection.Identity())
	}

	var wg sync.WaitGroup
	wg.Add(30)

	update := func() {
		for i := 1; i <= 10; i++ {
			connection.UpdateBytesIn(100)
			connection.UpdateBytesOut(100)
			wg.Done()
		}
	}
	go update()
	go update()
	go update()

	wg.Wait()
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

func TestSortByIdenity(t *testing.T) {

	ab := NewConnectionInfo("A", "B").Get()
	cd := NewConnectionInfo("C", "D").Get()
	ef := NewConnectionInfo("E", "F").Get()

	all := []*ConnectionInfo{ef, ab, cd}

	sort.Sort(ByIdentity(all))

	if all[0].Identity() != ab.Identity() {
		t.Error("Not sorted by Identity")
	}

}

func TestFilterByProperty(t *testing.T) {

	ab := NewConnectionInfo("A", "B").Get()
	cd := NewConnectionInfo("C", "D").Get()
	ef := NewConnectionInfo("E", "F").Get()

	all := &ConnectionInfoCollection{ef, ab, cd}

	filter := func(ci *ConnectionInfo) bool {
		return ci.To == "B"
	}

	filtered := all.Select(filter)

	if len(filtered) != 1 {
		t.Errorf("Incorrect number returned : %d", len(filtered))
	}

	if filtered[0].Identity() != "A:B" {
		t.Errorf("Incorrect connection returned by Identity() : %s", filtered[0].Identity())
	}

}
