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

// TestBindingPropagation verifies that Binding → Pod nodeName propagation works
func TestBindingPropagation(t *testing.T) {
    sim, err := New()
    if err != nil {
        t.Fatalf("Failed to create simulator: %v", err)
    }
    defer sim.Close()
    s := sim.(*Simulator) // type assertion to access unexported fields

    // Create a test Pod
    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-pod",
            Namespace: "default",
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
        },
    }

    _, err = s.client.CoreV1().Pods("default").Create(s.ctx, pod, metav1.CreateOptions{})
    if err != nil {
        t.Fatalf("Failed to create pod: %v", err)
    }

    // Create a Binding
    binding := &corev1.Binding{
        ObjectMeta: metav1.ObjectMeta{Name: pod.Name, Namespace: "default"},
        Target:     corev1.ObjectReference{Kind: "Node", Name: "test-node"},
    }
    err = s.client.CoreV1().Pods("default").Bind(s.ctx, binding, metav1.CreateOptions{})
    if err != nil {
        t.Fatalf("Failed to create binding: %v", err)
    }

    // Wait for propagation
    timeout := time.After(5 * time.Second)
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-timeout:
            t.Fatal("Timeout: Pod nodeName was not updated")
        case <-ticker.C:
            updated, _ := s.client.CoreV1().Pods("default").Get(s.ctx, pod.Name, metav1.GetOptions{})
            if updated != nil && updated.Spec.NodeName == "test-node" {
                t.Logf("✅ Pod successfully bound to node %s", updated.Spec.NodeName)
                return
            }
        }
    }
}

// TestLargeScaleScheduling stresses the simulator with thousands of pods
func TestLargeScaleScheduling(t *testing.T) {
    sim, err := New()
    if err != nil {
        t.Fatalf("Failed to create simulator: %v", err)
    }
    defer sim.Close()
    s := sim.(*Simulator)

    // Create 50 nodes
    for i := 0; i < 50; i++ {
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

    // Create 1000 pods
    podCount := 1000
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

    if len(failedPods) > podCount/10 {
        t.Errorf("Too many failed pods: %d/%d", len(failedPods), podCount)
    }
}

// TestConcurrentScheduling detects race conditions
func TestConcurrentScheduling(t *testing.T) {
    sim, err := New()
    if err != nil {
        t.Fatalf("Failed to create simulator: %v", err)
    }
    defer sim.Close()
    s := sim.(*Simulator)

    // Create nodes
    for i := 0; i < 10; i++ {
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

    podCount := 100
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

// TestNoGoroutineLeaks detects goroutine leaks
func TestNoGoroutineLeaks(t *testing.T) {
    initial := runtime.NumGoroutine()

    sim, err := New()
    if err != nil {
        t.Fatalf("Failed to create simulator: %v", err)
    }
    s := sim.(*Simulator)

    for i := 0; i < 100; i++ {
        pod := &corev1.Pod{
            ObjectMeta: metav1.ObjectMeta{
                Name:      fmt.Sprintf("leak-test-%d", i),
                Namespace: "default",
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
            },
        }
        s.createPod(pod)
        s.deletePod(pod)
    }

    sim.Close()
    time.Sleep(2 * time.Second)

    final := runtime.NumGoroutine()
    if final > initial+5 {
        t.Errorf("Potential goroutine leak: initial=%d, final=%d", initial, final)
    }
}

// TestEdgeCases tests various error scenarios
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

        pod := &corev1.Pod{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "no-node-pod",
                Namespace: "default",
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
                NodeName:   "nonexistent-node",
            },
        }

        err = s.createPod(pod)
        // This should fail gracefully
        t.Logf("Create pod with nonexistent node result: %v", err)
    })
}