// /*
// Copyright 2021 The Kubernetes Authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

// package scheduler

// import (
// 	"k8s.io/kubernetes/pkg/scheduler/framework"
// 	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/imagelocality"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/interpodaffinity"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/nodeaffinity"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/nodename"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/noderesources"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/nodeunschedulable"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/podtopologyspread"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/tainttoleration"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/volumerestrictions"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/volumezone"

// 	// ================================================================
// 	// H‑TAFM PLUGIN REGISTRATION (BLANK IMPORT)
// 	// This triggers the init() function in the htafm package,
// 	// which registers the "HTAFMScore" plugin with the runtime.
// 	// ================================================================
// 	// _ "github.com/anuxlab/anu_kubernetes-scheduler-simulator/pkg/scheduler/htafm"
// 	_ "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/scheduler/htafm"
// 	// Custom simulator plugins (from the HKUST-ADSL fork)
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/bestfit"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/dotproduct"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/fgd"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/gpuclustering"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/gpupacking"
// 	"k8s.io/kubernetes/pkg/scheduler/plugins/random"
// )

// // RegisterAllPlugins registers all built-in scheduler plugins.
// // This is called by the scheduler framework during initialization.
// func RegisterAllPlugins(registry *runtime.Registry) {
// 	// ------------------------------------------------------------------
// 	// Kubernetes core plugins
// 	// ------------------------------------------------------------------
// 	registry.Register("ImageLocality", imagelocality.New)
// 	registry.Register("InterPodAffinity", interpodaffinity.New)
// 	registry.Register("NodeAffinity", nodeaffinity.New)
// 	registry.Register("NodeName", nodename.New)
// 	registry.Register("NodeResourcesBalancedAllocation", noderesources.NewBalancedAllocation)
// 	registry.Register("NodeResourcesLeastAllocated", noderesources.NewLeastAllocated)
// 	registry.Register("NodeUnschedulable", nodeunschedulable.New)
// 	registry.Register("PodTopologySpread", podtopologyspread.New)
// 	registry.Register("TaintToleration", tainttoleration.New)
// 	registry.Register("VolumeRestrictions", volumerestrictions.New)
// 	registry.Register("VolumeZone", volumezone.New)

// 	// ------------------------------------------------------------------
// 	// Simulator custom plugins (from HKUST-ADSL)
// 	// ------------------------------------------------------------------
// 	registry.Register("BestFit", bestfit.New)
// 	registry.Register("DotProduct", dotproduct.New)
// 	registry.Register("FGD", fgd.New)
// 	registry.Register("GpuClustering", gpuclustering.New)
// 	registry.Register("GpuPacking", gpupacking.New)
// 	registry.Register("Random", random.New)

// 	// ------------------------------------------------------------------
// 	// H‑TAFM plugin is automatically registered via its init() function,
// 	// which is triggered by the blank import above.
// 	// The plugin's registration looks like:
// 	//   func init() { runtime.Register("HTAFMScore", htafm.New) }
// 	// ------------------------------------------------------------------
// }

// // GetDefaultPlugins returns the default set of plugins used by the simulator.
// // This function is used in the scheduler configuration generation.
// func GetDefaultPlugins() *framework.Plugins {
// 	return &framework.Plugins{
// 		Bind: framework.PluginSet{
// 			Disabled: []framework.Plugin{{Name: "DefaultBinder"}},
// 			Enabled:  []framework.Plugin{{Name: "Simon"}},
// 		},
// 		Filter: framework.PluginSet{
// 			Enabled: []framework.Plugin{{Name: "Open-Gpu-Share"}},
// 		},
// 		Score: framework.PluginSet{
// 			Disabled: []framework.Plugin{
// 				{Name: "RandomScore"},
// 				{Name: "DotProductScore"},
// 				{Name: "GpuClusteringScore"},
// 				{Name: "GpuPackingScore"},
// 				{Name: "BestFitScore"},
// 				{Name: "FGDScore"},
// 				{Name: "ImageLocality"},
// 				{Name: "NodeAffinity"},
// 				{Name: "PodTopologySpread"},
// 				{Name: "TaintToleration"},
// 				{Name: "NodeResourcesBalancedAllocation"},
// 				{Name: "InterPodAffinity"},
// 				{Name: "NodeResourcesLeastAllocated"},
// 				{Name: "NodePreferAvoidPods"},
// 			},
// 			// The actual enabled set is configured dynamically via the scheduler config
// 			// (e.g., FGDScore, BestFitScore, HTAFMScore, etc.)
// 			Enabled: []framework.Plugin{},
// 		},
// 		Reserve: framework.PluginSet{
// 			Enabled: []framework.Plugin{{Name: "Open-Gpu-Share"}},
// 		},
// 	}
// }


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

    // H-TAFM PLUGIN REGISTRATION
    // _ "github.com/anuxlab/anu_kubernetes-scheduler-simulator/pkg/scheduler/htafm"
    _ "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/scheduler/htafm"

    // Custom simulator plugins
    "k8s.io/kubernetes/pkg/scheduler/plugins/bestfit"
    "k8s.io/kubernetes/pkg/scheduler/plugins/dotproduct"
    "k8s.io/kubernetes/pkg/scheduler/plugins/fgd"
    "k8s.io/kubernetes/pkg/scheduler/plugins/gpuclustering"
    "k8s.io/kubernetes/pkg/scheduler/plugins/gpupacking"
    "k8s.io/kubernetes/pkg/scheduler/plugins/random"
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
}