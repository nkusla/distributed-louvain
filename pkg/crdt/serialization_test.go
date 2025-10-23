package crdt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNodeSetJSONSerialization(t *testing.T) {
	nodeSet := NewNodeSet()
	nodeSet.Add(1, 2, 0.5)
	nodeSet.Add(3, 4, 0.7)
	nodeSet.Add(5, 6, 0.3)
	nodeSet.Add(7, 8, 0.9)

	t.Logf("Created NodeSet with %d transitions", nodeSet.Size())

	jsonData, err := json.MarshalIndent(nodeSet, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal NodeSet: %v", err)
	}

	t.Logf("Serialized JSON: %s", string(jsonData))

	path := "../test/crdt/nodeset.json"
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write JSON to file: %v", err)
	}

	t.Logf("JSON written to %s", path)

	restored := NewNodeSet()
	err = json.Unmarshal(jsonData, restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal NodeSet: %v", err)
	}

	if restored.Size() != nodeSet.Size() {
		t.Errorf("Size mismatch: expected %d, got %d", nodeSet.Size(), restored.Size())
	}

	transition, exists := restored.Get(2)
	if !exists {
		t.Error("Expected transition for node 2 to exist")
	} else if transition.ModularityDelta != 0.5 {
		t.Errorf("Expected modularity delta 0.5 for node 2, got %f", transition.ModularityDelta)
	}

	transition, exists = restored.Get(8)
	if !exists {
		t.Error("Expected transition for node 8 to exist")
	} else if transition.ModularityDelta != 0.9 {
		t.Errorf("Expected modularity delta 0.9 for node 8, got %f", transition.ModularityDelta)
	}

	t.Log("JSON serialization test completed successfully!")
}
