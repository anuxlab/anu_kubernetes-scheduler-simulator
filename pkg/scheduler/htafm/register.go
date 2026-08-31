package htafm

import (
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func init() {
	runtime.Register(Name, New)
}