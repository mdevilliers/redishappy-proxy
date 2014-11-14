package proxy

import (
	"testing"
)

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
