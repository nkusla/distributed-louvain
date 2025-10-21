package graph

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
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
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	edges := make([]Edge, 0, len(records))
	for i, record := range records {
		if len(record) != 3 {
			return nil, fmt.Errorf("line %d: expected 3 columns, got %d", i+1, len(record))
		}

		u, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid source node: %w", i+1, err)
		}

		v, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid target node: %w", i+1, err)
		}

		w, err := strconv.Atoi(record[2])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid weight: %w", i+1, err)
		}

		edges = append(edges, NewEdge(u, v, w))
		edges = append(edges, NewEdge(v, u, w))
	}

	return edges, nil
}