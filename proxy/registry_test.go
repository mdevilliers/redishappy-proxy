package proxy

import "testing"

func TestBasicRegistryUsage(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterConnection("A", "B", &Proxy{})

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
	registry.RegisterConnection("A", "B", &Proxy{})
	registry.RegisterConnection("A", "C", &Proxy{})
	registry.RegisterConnection("A", "D", &Proxy{})

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

func TestStatistics(t *testing.T) {

	registry := NewRegistry()

	one := registry.RegisterConnection("A", "B", &Proxy{})
	registry.RegisterConnection("A", "C", &Proxy{})
	registry.RegisterConnection("A", "D", &Proxy{})

	stats := registry.GetStatistics()

	if stats.TotalNumberOfConnections != 3 {
		t.Errorf("Total number of Connections should be 3 but are %d", stats.TotalNumberOfConnections)
	}

	if stats.CurrentNumberOfConnections != 3 {
		t.Errorf("Total number of Current Connections should be 3 but are %d", stats.CurrentNumberOfConnections)
	}

	registry.UnRegisterConnection(one.Identity())

	stats = registry.GetStatistics()
	if stats.TotalNumberOfConnections != 3 {
		t.Errorf("Total number of Connections should be 3 but are %d", stats.TotalNumberOfConnections)
	}

	if stats.CurrentNumberOfConnections != 2 {
		t.Errorf("Total number of Current Connections should be 2 but are %d", stats.CurrentNumberOfConnections)
	}
}

func TestUpdatingConnectionInRegistry(t *testing.T) {
	registry := NewRegistry()

	one := registry.RegisterConnection("A", "B", &Proxy{})

	registry.UpdateExistingConnection(one.Identity(), "C")

	if one.to != "C" {

		t.Errorf("Existing connection should be pointing to 'C' not %s", one.to)
	}

	registry.UpdateExistingConnection(one.Identity(), "D")

	if one.to != "D" {

		t.Errorf("Existing connection should be pointing to 'D' not %s", one.to)
	}
}
