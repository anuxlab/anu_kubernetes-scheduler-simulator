package htafm

import (
	"context"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const Name = "HTAFMScore"

type HTAFM struct {
	handle  framework.Handle
	variant string // "cut", "entropy", "hier"
	hg      *Hypergraph
}

var _ framework.ScorePlugin = &HTAFM{}

// Args defines the plugin configuration.
type Args struct {
	Variant string `json:"variant"`
}

// New initializes the plugin with the provided arguments.
func New(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
	args := &Args{}
	if obj != nil {
		if err := runtime.DecodeInto(/* codec */, obj, args); err != nil {
			return nil, err
		}
	}
	// Default to "cut" if not specified
	if args.Variant == "" {
		args.Variant = "cut"
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
	node := nodeInfo.Node()

	// Build hypergraph (in production, you'd maintain a global one)
	allNodes, _ := pl.handle.SnapshotSharedLister().NodeInfos().List()
	nodes := make([]*v1.Node, 0, len(allNodes))
	for _, ni := range allNodes {
		nodes = append(nodes, ni.Node())
	}
	pl.hg = NewHypergraph(nodes)

	// Compute current TAFI (as a proxy)
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