package htafm

import (
	"math"

	"k8s.io/api/core/v1"
)

// ComputeTAFI computes the fragmentation index for the current cluster state.
// It uses the hypergraph and the node resource usage.
func ComputeTAFI(hg *Hypergraph, nodes []*v1.Node, variant string) float64 {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	var totalTAFI float64
	for _, edge := range hg.Hyperedges {
		phi := computePhi(edge, nodes, variant)
		totalTAFI += edge.Weight * phi
	}
	return totalTAFI
}

func computePhi(edge *Hyperedge, nodes []*v1.Node, variant string) float64 {
	switch variant {
	case "cut":
		return cutPhi(edge, nodes)
	case "entropy":
		return entropyPhi(edge, nodes)
	case "hier":
		return hierarchicalPhi(edge, nodes)
	default:
		return cutPhi(edge, nodes)
	}
}

// cutPhi returns 1 if the aggregate free capacity in the hyperedge is sufficient
// for some pending VM, but no single vertex can accommodate it.
func cutPhi(edge *Hyperedge, nodes []*v1.Node) float64 {
	// Aggregate free capacity across vertices
	var totalCPU, totalMem, totalGPU int64
	var maxCPU, maxMem, maxGPU int64
	for _, vID := range edge.Vertices {
		node := findNode(nodes, vID)
		if node == nil {
			continue
		}
		freeCPU := node.Status.Allocatable.Cpu().MilliValue() - getUsedCPU(node)
		freeMem := node.Status.Allocatable.Memory().Value() - getUsedMemory(node)
		freeGPU := getGPUResource(node) - getUsedGPU(node)

		totalCPU += freeCPU
		totalMem += freeMem
		totalGPU += freeGPU
		if freeCPU > maxCPU {
			maxCPU = freeCPU
		}
		if freeMem > maxMem {
			maxMem = freeMem
		}
		if freeGPU > maxGPU {
			maxGPU = freeGPU
		}
	}
	// Here we assume a pending VM of size (1 CPU, 1GB, 1GPU) as a threshold.
	// In reality you'd check against actual pending pods.
	// For simplicity, we use a heuristic: if total > 10% of capacity but no single vertex can hold 20% of the total.
	if totalCPU > 0 && totalMem > 0 && totalGPU > 0 {
		// If aggregate is enough to schedule a large VM but each vertex is too small.
		if totalCPU > 1000 && totalMem > 1e9 && totalGPU > 0 {
			if maxCPU < totalCPU/2 && maxMem < totalMem/2 && maxGPU < totalGPU/2 {
				return 1.0
			}
		}
	}
	return 0.0
}

// entropyPhi computes the entropy of free resource distribution across vertices.
func entropyPhi(edge *Hyperedge, nodes []*v1.Node) float64 {
	var freeVec []float64
	var totalFree float64
	for _, vID := range edge.Vertices {
		node := findNode(nodes, vID)
		if node == nil {
			continue
		}
		// Normalised free capacity: sum of CPU, memory, GPU fractions
		cpuFrac := float64(node.Status.Allocatable.Cpu().MilliValue()-getUsedCPU(node)) / float64(node.Status.Allocatable.Cpu().MilliValue())
		memFrac := float64(node.Status.Allocatable.Memory().Value()-getUsedMemory(node)) / float64(node.Status.Allocatable.Memory().Value())
		gpuFrac := float64(getGPUResource(node)-getUsedGPU(node)) / float64(getGPUResource(node)+1) // avoid div by zero
		norm := cpuFrac + memFrac + gpuFrac
		freeVec = append(freeVec, norm)
		totalFree += norm
	}
	if totalFree == 0 {
		return 0.0
	}
	var entropy float64
	for _, f := range freeVec {
		p := f / totalFree
		if p > 0 {
			entropy -= p * math.Log(p)
		}
	}
	return entropy
}

// hierarchicalPhi considers fragmentation at multiple topological levels.
func hierarchicalPhi(edge *Hyperedge, nodes []*v1.Node) float64 {
	thresholds := map[string]float64{
		"NUMA":   0.5,
		"Socket": 0.4,
		"Server": 0.3,
		"Rack":   0.2,
	}
	level := edge.Level
	threshold, ok := thresholds[level]
	if !ok {
		threshold = 0.3
	}
	fragCount := 0
	for _, vID := range edge.Vertices {
		node := findNode(nodes, vID)
		if node == nil {
			continue
		}
		freeCPU := float64(node.Status.Allocatable.Cpu().MilliValue()-getUsedCPU(node)) / float64(node.Status.Allocatable.Cpu().MilliValue())
		freeMem := float64(node.Status.Allocatable.Memory().Value()-getUsedMemory(node)) / float64(node.Status.Allocatable.Memory().Value())
		freeGPU := float64(getGPUResource(node)-getUsedGPU(node)) / float64(getGPUResource(node)+1)
		if freeCPU < threshold || freeMem < threshold || freeGPU < threshold {
			fragCount++
		}
	}
	if len(edge.Vertices) == 0 {
		return 0.0
	}
	return float64(fragCount) / float64(len(edge.Vertices))
}

// Helper functions to retrieve used resources from node status.
// In practice you'd get them from the pod list.
func getUsedCPU(node *v1.Node) int64 { return 0 }   // placeholder
func getUsedMemory(node *v1.Node) int64 { return 0 } // placeholder
func getUsedGPU(node *v1.Node) int64 { return 0 }   // placeholder

func findNode(nodes []*v1.Node, name string) *v1.Node {
	for _, n := range nodes {
		if n.Name == name {
			return n
		}
	}
	return nil
}