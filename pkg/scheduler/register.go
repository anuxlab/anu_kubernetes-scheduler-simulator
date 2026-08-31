package scheduler

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	"k8s.io/kubernetes/pkg/scheduler/plugins/imagelocality"
	"k8s.io/kubernetes/pkg/scheduler/plugins/interpodaffinity"
	"k8s.io/kubernetes/pkg/scheduler/plugins/nodeaffinity"
	"k8s.io/kubernetes/pkg/scheduler/plugins/nodename"
	"k8s.io/kubernetes/pkg/scheduler/plugins/noderesources"
	"k8s.io/kubernetes/pkg/scheduler/plugins/nodeunschedulable"
	"k8s.io/kubernetes/pkg/scheduler/plugins/podtopologyspread"
	"k8s.io/kubernetes/pkg/scheduler/plugins/tainttoleration"
	"k8s.io/kubernetes/pkg/scheduler/plugins/volumerestrictions"
	"k8s.io/kubernetes/pkg/scheduler/plugins/volumezone"

	// Simulator custom plugins
	"k8s.io/kubernetes/pkg/scheduler/plugins/bestfit"
	"k8s.io/kubernetes/pkg/scheduler/plugins/dotproduct"
	"k8s.io/kubernetes/pkg/scheduler/plugins/fgd"
	"k8s.io/kubernetes/pkg/scheduler/plugins/gpuclustering"
	"k8s.io/kubernetes/pkg/scheduler/plugins/gpupacking"
	"k8s.io/kubernetes/pkg/scheduler/plugins/random"

	// Blank import to trigger the init() of the H‑TAFM plugin
	_ "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/scheduler/htafm"
)

func RegisterAllPlugins(registry *runtime.Registry) {
	// Core Kubernetes plugins
	registry.Register("ImageLocality", imagelocality.New)
	registry.Register("InterPodAffinity", interpodaffinity.New)
	registry.Register("NodeAffinity", nodeaffinity.New)
	registry.Register("NodeName", nodename.New)
	registry.Register("NodeResourcesBalancedAllocation", noderesources.NewBalancedAllocation)
	registry.Register("NodeResourcesLeastAllocated", noderesources.NewLeastAllocated)
	registry.Register("NodeUnschedulable", nodeunschedulable.New)
	registry.Register("PodTopologySpread", podtopologyspread.New)
	registry.Register("TaintToleration", tainttoleration.New)
	registry.Register("VolumeRestrictions", volumerestrictions.New)
	registry.Register("VolumeZone", volumezone.New)

	// Simulator custom plugins
	registry.Register("BestFit", bestfit.New)
	registry.Register("DotProduct", dotproduct.New)
	registry.Register("FGD", fgd.New)
	registry.Register("GpuClustering", gpuclustering.New)
	registry.Register("GpuPacking", gpupacking.New)
	registry.Register("Random", random.New)

    registry.Register("HTAFMScore", htafm.New)

	// The H‑TAFM plugin is registered via its own init() function,
	// triggered by the blank import above.
}