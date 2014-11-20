package acceptor

import (
	"testing"

	"github.com/mdevilliers/redishappy-proxy/proxy"
)

func TestBasicPoolUsage(t *testing.T) {

	pool := NewAcceptorPool(proxy.NewRegistry())

	acceptor, err := pool.NewOrDefaultAcceptor("test", "1.1.1.1:8080", "2.2.2.2:8080")

	if err != nil {
		t.Error(" pool.NewOrDefaultAcceptor should not error")
	}

	if acceptor == nil {
		t.Error("Acceptor cannot be nil")
	}

	acceptor2, err := pool.NewOrDefaultAcceptor("test", "3.3.3.3:8080", "4.4.4.4:8080")

	if err != nil {
		t.Error(" pool.NewOrDefaultAcceptor should not error")
	}

	if acceptor != acceptor2 {
		t.Error("Acceptors should have the same reference if called the same name.")
	}
}

func TestInvalidAcceptor(t *testing.T) {

	pool := NewAcceptorPool(proxy.NewRegistry())

	_, err := pool.NewOrDefaultAcceptor("test", "rubbish", "2.2.2.2:8080")

	if err == nil {
		t.Error(" pool.NewOrDefaultAcceptor should error as passed incorrect address")
	}
}

func TestRemovingExistingAcceptor(t *testing.T) {
	pool := NewAcceptorPool(proxy.NewRegistry())

	name := "test"
	from := "1.1.1.1:8080"
	to := "2.2.2.2:8080"

	pool.NewOrDefaultAcceptor(name, from, to)

	pool.RemoveExistingAcceptor(name)

	_, found := pool.m[name]
	if found {
		t.Error("Acceptor should have been removed.")
	}
}

func TestReplaceAcceptor(t *testing.T) {

	pool := NewAcceptorPool(proxy.NewRegistry())

	name := "test"

	_, err := pool.NewOrDefaultAcceptor(name, "1.1.1.1:8080", "2.2.2.2:8080")

	if err != nil {
		t.Error(" pool.NewOrDefaultAcceptor should not error")
	}

	_, err = pool.ReplaceOrDefaultAcceptor(name, "3.3.3.3:8080", "4.4.4.4:8080")

	if err != nil {
		t.Error(" pool.ReplaceOrDefaultAcceptor should not error")
	}

	if len(pool.m) != 1 {
		t.Errorf("Pool should only contain one member. It contains %d", len(pool.m))
	}

}
