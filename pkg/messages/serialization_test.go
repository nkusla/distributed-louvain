package messages

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/crdt"
	"github.com/distributed-louvain/pkg/graph"
)

func TestMessageJSONSerialization(t *testing.T) {
	testPID := actor.NewPID("node-1", "partition-0")
	testEdges := []graph.Edge{
		{U: 1, V: 2, W: 1},
		{U: 2, V: 3, W: 2},
		{U: 3, V: 4, W: 1},
	}
	testNodeSet := crdt.NewNodeSet()
	testNodeSet.Add(1, 2, 0.5)
	testNodeSet.Add(3, 4, 0.7)

	testCases := []struct {
		name    string
		message interface{}
	}{
		{
			name: "InitialPartitionCreation",
			message: &InitialPartitionCreation{
				Edges:            testEdges,
				TotalGraphWeight: 78,
			},
		},
		{
			name: "InitialPartitionCreationComplete",
			message: &InitialPartitionCreationComplete{
				Sender: testPID,
			},
		},
		{
			name: "StartPhase1",
			message: &StartPhase1{},
		},
		{
			name: "DegreeRequest",
			message: &DegreeRequest{
				NodeID:     1,
				NeighborID: 2,
				Sender:    testPID,
			},
		},
		{
			name: "DegreeResponse",
			message: &DegreeResponse{
				NodeID:     1,
				NeighborID: 2,
				Degree:     5,
			},
		},
		{
			name: "LocalOptimizationComplete",
			message: &LocalOptimizationComplete{
				NodeSet: testNodeSet,
				Sender:  testPID,
			},
		},
		{
			name: "StartPhase2",
			message: &StartPhase2{},
		},
		{
			name: "EdgeAggregate",
			message: &EdgeAggregate{
				CommunityU: 1,
				CommunityV: 2,
				Weight:     3,
				Sender:     testPID,
			},
		},
		{
			name: "EdgeAggregateComplete",
			message: &EdgeAggregateComplete{
				Sender: testPID,
			},
		},
		{
			name: "AggregationResult",
			message: &AggregationResult{
				Edges: testEdges,
			},
		},
		{
			name: "AggregationComplete",
			message: &AggregationComplete{
				Sender: testPID,
			},
		},
		{
			name: "CompleteAggregation",
			message: &CompleteAggregation{},
		},
		{
			name: "AlgorithmComplete",
			message: &AlgorithmComplete{
				FinalModularity: 0.123456,
				Iterations:      5,
			},
		},
		{
			name: "Shutdown",
			message: &Shutdown{},
		},
	}

	path := "../test/messages"
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.MarshalIndent(tc.message, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal %s: %v", tc.name, err)
			}

			t.Logf("Serialized %s: %s", tc.name, string(jsonData))

			filename := filepath.Join(path, tc.name+".json")
			err = os.WriteFile(filename, jsonData, 0644)
			if err != nil {
				t.Fatalf("Failed to write %s to file: %v", tc.name, err)
			}

			if tc.name != "StartPhase1" && tc.name != "StartPhase2" &&
			   tc.name != "CompleteAggregation" && tc.name != "Shutdown" {

				var restored interface{}
				switch tc.name {
				case "InitialPartitionCreation":
					restored = &InitialPartitionCreation{}
				case "InitialPartitionCreationComplete":
					restored = &InitialPartitionCreationComplete{}
				case "DegreeRequest":
					restored = &DegreeRequest{}
				case "DegreeResponse":
					restored = &DegreeResponse{}
				case "LocalOptimizationComplete":
					restored = &LocalOptimizationComplete{}
				case "EdgeAggregate":
					restored = &EdgeAggregate{}
				case "EdgeAggregateComplete":
					restored = &EdgeAggregateComplete{}
				case "AggregationResult":
					restored = &AggregationResult{}
				case "AggregationComplete":
					restored = &AggregationComplete{}
				case "AlgorithmComplete":
					restored = &AlgorithmComplete{}
				}

				err = json.Unmarshal(jsonData, restored)
				if err != nil {
					t.Fatalf("Failed to unmarshal %s: %v", tc.name, err)
				}

				t.Logf("Successfully deserialized %s", tc.name)
			}
		})
	}

	t.Logf("All message serialization tests completed! Output files written to: %s", path)
}
