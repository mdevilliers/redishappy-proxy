package proxy

import "testing"

func TestBasicRegistryUsage(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterConnection("A", "B")

	connections := registry.GetConnections()

	if len(connections) != 1 {
		t.Errorf("There should only be one connection. There are %d", len(connections))
	}

	registry.UnRegisterConnection("A:B")

	connections = registry.GetConnections()

	if len(connections) != 0 {
		t.Errorf("There should be no connections. There are %d", len(connections))
	}

}

func TestFilter(t *testing.T) {

	registry := NewRegistry()
	registry.RegisterConnection("A", "B")
	registry.RegisterConnection("A", "C")
	registry.RegisterConnection("A", "D")

	filter := func(ci *ConnectionInfo) bool {
		return ci.To == "B"
	}

	results := registry.GetConnectionsWithFilter(filter)

	if len(results) != 1 {
		t.Errorf("There should be 1 result. There are %d", len(results))
	}

	filter = func(ci *ConnectionInfo) bool {
		return ci.To == "B" || ci.To == "C"
	}

	results = registry.GetConnectionsWithFilter(filter)

	if len(results) != 2 {
		t.Errorf("There should be 2 results. There are %d", len(results))
	}
}
