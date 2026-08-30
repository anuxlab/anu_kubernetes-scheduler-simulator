package htafm

import (
	_ "github.com/anuxlab/anu_kubernetes-scheduler-simulator/pkg/scheduler/htafm" // blank import (optional – only needed if you want to trigger init from outside)
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func init() {
	// Register the H‑TAFM plugin with the scheduler framework.
	// The framework will call New() with the appropriate configuration.
	runtime.Register(Name, New)
}