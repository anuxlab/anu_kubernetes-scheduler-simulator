package htafm

import (
	"context"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const Name = "HTAFMScore"

type HTAFM struct {
	handle  framework.Handle
	variant string
	hg      *Hypergraph
}

var _ framework.ScorePlugin = &HTAFM{}

type Args struct {
	Variant string `json:"variant"`
}

// New initializes the plugin with the provided arguments.
func New(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
	args := &Args{}
	// Default variant
	args.Variant = "cut"

	if obj != nil {
		// The object is an unstructured.Unstructured containing the config
		if u, ok := obj.(*unstructured.Unstructured); ok {
			if variant, found, err := unstructured.NestedString(u.Object, "variant"); err == nil && found {
				args.Variant = variant
			}
		} else {
			// Fallback: try to decode via JSON (if needed)
			// For simplicity, we just use the default.
		}
	}

	return &HTAFM{
		handle:  h,
		variant: args.Variant,
	}, nil
}

func (pl *HTAFM) Name() string {
	return Name
}

// Score computes the fragmentation score for a node.
func (pl *HTAFM) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// Get node info
	nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, "failed to get node info")
	}
	_ = nodeInfo.Node() // node not used directly; we use the list of all nodes

	// Build hypergraph from all nodes
	allNodes, _ := pl.handle.SnapshotSharedLister().NodeInfos().List()
	nodes := make([]*v1.Node, 0, len(allNodes))
	for _, ni := range allNodes {
		nodes = append(nodes, ni.Node())
	}
	pl.hg = NewHypergraph(nodes)

	// Compute current TAFI
	tafi := ComputeTAFI(pl.hg, nodes, pl.variant)

	// Score = 100 - tafi (lower fragmentation = higher score)
	score := int64(100 - tafi*100)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score, framework.NewStatus(framework.Success, "")
}

func (pl *HTAFM) ScoreExtensions() framework.ScoreExtensions {
	return nil
}