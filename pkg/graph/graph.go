package graph

import (
	"github.com/distributed-louvain/pkg/graphio"
)

type Edge struct {
	U   int
	V   int
	W   int
}

func NewEdge(u, v int, w int) Edge {
	return Edge{U: u, V: v, W: w}
}

type Neighbor struct {
	NodeID   int
	Weight int
}

type Graph struct {
	Adj    map[int][]Neighbor
	Degree map[int]int
}

func NewGraph() *Graph {
	return &Graph{
		Adj:    make(map[int][]Neighbor),
		Degree: make(map[int]int),
	}
}

func (g *Graph) AddEdge(edge Edge) {
	if edge.U == edge.V {
		return
	}

	g.Adj[edge.U] = append(g.Adj[edge.U], Neighbor{NodeID: edge.V, Weight: edge.W})
	g.Adj[edge.V] = append(g.Adj[edge.V], Neighbor{NodeID: edge.U, Weight: edge.W})

	g.Degree[edge.U] += edge.W
	g.Degree[edge.V] += edge.W
}

func (g *Graph) AddEdges(edges []Edge) {
	for _, edge := range edges {
		g.AddEdge(edge)
	}
}

func (g *Graph) GetWeight(nodeU, nodeV int) int {
	neighbors, exists := g.Adj[nodeU]
	if !exists {
		return 0
	}

	for _, neighbor := range neighbors {
		if neighbor.NodeID == nodeV {
			return neighbor.Weight
		}
	}

	return 0
}

func ReadEdgesFromCSV(filename string) ([]Edge, error) {
	records, err := graphio.ReadCSVWithHeader(filename, true, "u")
	if err != nil {
		return nil, err
	}

	edges := make([]Edge, 0, len(records))
	for i, record := range records {
		lineNum := i + 2 // +2 because we account for header and 1-based indexing

		if err := graphio.ValidateRecordLength(record, 3, lineNum); err != nil {
			return nil, err
		}

		values, err := graphio.ParseIntRecord(record, lineNum)
		if err != nil {
			return nil, err
		}

		edges = append(edges, NewEdge(values[0], values[1], values[2]))
	}

	return edges, nil
}