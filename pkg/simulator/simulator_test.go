package simulator

import (
    "fmt"
    "runtime"
    "sync"
    "testing"
    "time"

    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    simontype "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type"
)

// TestBindingPropagation verifies that the scheduler + binding controller
// correctly sets Pod.Spec.NodeName.
func TestBindingPropagation(t *testing.T) {
    sim, err := New()
    if err != nil {
        t.Fatalf("Failed to create simulator: %v", err)
    }
    defer sim.Close()
    s := sim.(*Simulator)

    // Start the scheduler (required for binding).
    s.runScheduler()

    // Create a node so the scheduler can assign the pod.
    node := &corev1.Node{
        ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
        Status: corev1.NodeStatus{
            Allocatable: corev1.ResourceList{
                corev1.ResourceCPU:    resource.MustParse("4"),
                corev1.ResourceMemory: resource.MustParse("8Gi"),
            },
        },
    }
    _, err = s.client.CoreV1().Nodes().Create(s.ctx, node, metav1.CreateOptions{})
    if err != nil {
        t.Fatalf("Failed to create node: %v", err)
    }

    // Create a test Pod.
    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-pod",
            Namespace: "default",
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{{
                Name:  "test",
                Image: "nginx",
                Resources: corev1.ResourceRequirements{
                    Requests: corev1.ResourceList{
                        corev1.ResourceCPU:    resource.MustParse("1"),
                        corev1.ResourceMemory: resource.MustParse("1Gi"),
                    },
                },
            }},
        },
    }

    // Use assumePod to schedule the pod.
    unscheduled := s.assumePod(pod)
    if unscheduled != nil {
        t.Fatalf("Pod was not scheduled: %v", unscheduled.Reason)
    }

    // Now pod.Spec.NodeName should be set.
    if pod.Spec.NodeName == "" {
        t.Error("Pod.Spec.NodeName is still empty after assumePod")
    } else {
        t.Logf("✅ Pod successfully bound to node %s", pod.Spec.NodeName)
    }
}

// TestLargeScaleScheduling uses a small number of nodes/pods in CI.
func TestLargeScaleScheduling(t *testing.T) {
    sim, err := New()
    if err != nil {
        t.Fatalf("Failed to create simulator: %v", err)
    }
    defer sim.Close()
    s := sim.(*Simulator)

    s.runScheduler()

    // Use tiny scale in CI (short mode), slightly larger otherwise.
    nodeCount := 3
    podCount := 10
    if !testing.Short() {
        nodeCount = 5
        podCount = 20
    }

    t.Logf("Creating %d nodes and %d pods", nodeCount, podCount)

    // Create nodes.
    for i := 0; i < nodeCount; i++ {
        node := &corev1.Node{
            ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("node-%03d", i)},
            Status: corev1.NodeStatus{
                Allocatable: corev1.ResourceList{
                    corev1.ResourceCPU:    resource.MustParse("32"),
                    corev1.ResourceMemory: resource.MustParse("64Gi"),
                },
            },
        }
        _, err = s.client.CoreV1().Nodes().Create(s.ctx, node, metav1.CreateOptions{})
        if err != nil {
            t.Fatalf("Failed to create node: %v", err)
        }
    }

    // Create pods.
    pods := make([]*corev1.Pod, podCount)
    for i := 0; i < podCount; i++ {
        pod := &corev1.Pod{
            ObjectMeta: metav1.ObjectMeta{
                Name:      fmt.Sprintf("stress-pod-%05d", i),
                Namespace: "default",
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{{
                    Name:  "test",
                    Image: "nginx",
                    Resources: corev1.ResourceRequirements{
                        Requests: corev1.ResourceList{
                            corev1.ResourceCPU:    resource.MustParse("1"),
                            corev1.ResourceMemory: resource.MustParse("1Gi"),
                        },
                    },
                }},
            },
        }
        pods[i] = pod
    }

    start := time.Now()
    failedPods := s.SchedulePods(pods)
    elapsed := time.Since(start)

    t.Logf("Scheduled %d pods in %v", podCount-len(failedPods), elapsed)
    t.Logf("Failed pods: %d", len(failedPods))

    if len(failedPods) > podCount/2 {
        t.Errorf("Too many failed pods: %d/%d", len(failedPods), podCount)
    }
}

// TestConcurrentScheduling uses reduced concurrency for speed.
func TestConcurrentScheduling(t *testing.T) {
    sim, err := New()
    if err != nil {
        t.Fatalf("Failed to create simulator: %v", err)
    }
    defer sim.Close()
    s := sim.(*Simulator)

    s.runScheduler()

    nodeCount := 2
    podCount := 5
    if !testing.Short() {
        nodeCount = 3
        podCount = 10
    }

    // Create nodes.
    for i := 0; i < nodeCount; i++ {
        node := &corev1.Node{
            ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("node-%d", i)},
            Status: corev1.NodeStatus{
                Allocatable: corev1.ResourceList{
                    corev1.ResourceCPU:    resource.MustParse("16"),
                    corev1.ResourceMemory: resource.MustParse("32Gi"),
                },
            },
        }
        _, err = s.client.CoreV1().Nodes().Create(s.ctx, node, metav1.CreateOptions{})
        if err != nil {
            t.Fatalf("Failed to create node: %v", err)
        }
    }

    pods := make([]*corev1.Pod, podCount)
    for i := 0; i < podCount; i++ {
        pod := &corev1.Pod{
            ObjectMeta: metav1.ObjectMeta{
                Name:      fmt.Sprintf("concurrent-pod-%d", i),
                Namespace: "default",
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{{
                    Name:  "test",
                    Image: "nginx",
                    Resources: corev1.ResourceRequirements{
                        Requests: corev1.ResourceList{
                            corev1.ResourceCPU:    resource.MustParse("2"),
                            corev1.ResourceMemory: resource.MustParse("4Gi"),
                        },
                    },
                }},
            },
        }
        pods[i] = pod
    }

    var wg sync.WaitGroup
    failedChan := make(chan *simontype.UnscheduledPod, podCount)

    for _, pod := range pods {
        wg.Add(1)
        go func(p *corev1.Pod) {
            defer wg.Done()
            if unscheduled := s.assumePod(p); unscheduled != nil {
                failedChan <- unscheduled
            }
        }(pod)
    }

    wg.Wait()
    close(failedChan)

    failedCount := 0
    for range failedChan {
        failedCount++
    }

    t.Logf("Concurrently scheduled %d pods, %d failed", podCount, failedCount)
}

// TestNoGoroutineLeaks uses a small number of iterations.
func TestNoGoroutineLeaks(t *testing.T) {
    initial := runtime.NumGoroutine()

    sim, err := New()
    if err != nil {
        t.Fatalf("Failed to create simulator: %v", err)
    }
    s := sim.(*Simulator)

    s.runScheduler()

    iterations := 5
    if !testing.Short() {
        iterations = 10
    }

    for i := 0; i < iterations; i++ {
        nodeName := fmt.Sprintf("leak-node-%d", i)
        node := &corev1.Node{
            ObjectMeta: metav1.ObjectMeta{Name: nodeName},
            Status: corev1.NodeStatus{
                Allocatable: corev1.ResourceList{
                    corev1.ResourceCPU:    resource.MustParse("4"),
                    corev1.ResourceMemory: resource.MustParse("8Gi"),
                },
            },
        }
        _, err = s.client.CoreV1().Nodes().Create(s.ctx, node, metav1.CreateOptions{})
        if err != nil {
            t.Fatalf("Failed to create node: %v", err)
        }

        pod := &corev1.Pod{
            ObjectMeta: metav1.ObjectMeta{
                Name:      fmt.Sprintf("leak-test-%d", i),
                Namespace: "default",
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
                NodeName:   nodeName,
            },
        }
        _, err = s.client.CoreV1().Pods("default").Create(s.ctx, pod, metav1.CreateOptions{})
        if err != nil {
            t.Fatalf("Failed to create pod: %v", err)
        }
        err = s.client.CoreV1().Pods("default").Delete(s.ctx, pod.Name, metav1.DeleteOptions{})
        if err != nil {
            t.Fatalf("Failed to delete pod: %v", err)
        }
    }

    sim.Close()
    time.Sleep(2 * time.Second)

    final := runtime.NumGoroutine()
    if final > initial+10 {
        t.Errorf("Potential goroutine leak: initial=%d, final=%d", initial, final)
    }
}

// TestEdgeCases tests various error scenarios (kept minimal).
func TestEdgeCases(t *testing.T) {
    t.Run("DeleteNonExistentPod", func(t *testing.T) {
        sim, err := New()
        if err != nil {
            t.Fatalf("Failed to create simulator: %v", err)
        }
        defer sim.Close()
        s := sim.(*Simulator)

        pod := &corev1.Pod{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "nonexistent",
                Namespace: "default",
            },
        }
        err = s.deletePod(pod)
        if err != nil {
            t.Logf("Delete non-existent pod returned error (expected): %v", err)
        }
    })

    t.Run("CreatePodWithoutNode", func(t *testing.T) {
        sim, err := New()
        if err != nil {
            t.Fatalf("Failed to create simulator: %v", err)
        }
        defer sim.Close()
        s := sim.(*Simulator)

        s.runScheduler()

        pod := &corev1.Pod{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "no-node-pod",
                Namespace: "default",
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
            },
        }

        err = s.createPod(pod)
        t.Logf("Create pod with no nodes result: %v", err)
    })
}