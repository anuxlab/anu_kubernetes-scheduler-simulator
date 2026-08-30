package htafm

import (
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// Vertex represents a physical resource unit (NUMA node, GPU, etc.)
type Vertex struct {
	ID       string
	NodeName string // associated Node name
	CPU      int64
	Memory   int64
	GPU      int64
	Level    string // "NUMA", "Socket", "Server", "Rack"
}

// Hyperedge represents a topological relationship (e.g., all vertices in the same server)
type Hyperedge struct {
	ID       string
	Level    string   // "NUMA", "Socket", "Server", "Rack", "Affinity"
	Vertices []string // Vertex IDs
	Weight   float64
}

// Hypergraph models the cluster topology
type Hypergraph struct {
	mu         sync.RWMutex
	Vertices   map[string]*Vertex
	Hyperedges map[string]*Hyperedge
}

// NewHypergraph builds the hypergraph from cluster nodes.
// In practice, you would read topology information from node labels.
func NewHypergraph(nodes []*v1.Node) *Hypergraph {
	hg := &Hypergraph{
		Vertices:   make(map[string]*Vertex),
		Hyperedges: make(map[string]*Hyperedge),
	}
	for _, node := range nodes {
		// Each node becomes a vertex (server level)
		vertex := &Vertex{
			ID:       node.Name,
			NodeName: node.Name,
			CPU:      node.Status.Capacity.Cpu().MilliValue(),
			Memory:   node.Status.Capacity.Memory().Value(),
			GPU:      getGPUResource(node), // assume helper function
			Level:    "Server",
		}
		hg.Vertices[node.Name] = vertex

		// Server‑level hyperedge (could include all vertices in the server)
		edge := &Hyperedge{
			ID:       "server-" + node.Name,
			Level:    "Server",
			Vertices: []string{node.Name},
			Weight:   0.5,
		}
		hg.Hyperedges[edge.ID] = edge

		// Additional topology can be added here (NUMA, sockets, racks)
		// based on node labels/annotations.
	}
	return hg
}

func getGPUResource(node *v1.Node) int64 {
	if val, ok := node.Status.Capacity["nvidia.com/gpu"]; ok {
		return val.Value()
	}
	return 0
}