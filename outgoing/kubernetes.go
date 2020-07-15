package outgoing

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	wylisconfig "github.com/chongyangshi/wylis/config"
)

type neighbourPod struct {
	podIP  string
	nodeIP string
}

// Normally, when calling pods of a Kubernetes service without using a separate
// service proxy, the client should call the cluster IP for kube-proxy to forward
// the reuqest to an arbitrary destination pod which could serve the it. We can't
// do that but instead must have a full picture of pods in the Wylis daemonset, as
// we need to reach pods in the daemonset running on all nodes.
var (
	clientSet        = &kubernetes.Clientset{}
	neighbourPodLock = sync.RWMutex{}
	neighbourPods    = []neighbourPod{}
)

func initClusterClient(ctx context.Context) error {
	// Load API client from default token
	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Error(ctx, "Could not load in-cluster config: %v", err)
		return err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error(ctx, "Could not load client set: %v", err)
		return err
	}

	clientSet = c
	err = refreshNeighbourPods(ctx)
	if err != nil {
		slog.Error(ctx, "Could not load neighbours: %v", err)
		return err
	}

	interval, err := time.ParseDuration(wylisconfig.ConfigRefreshInterval)
	if err != nil {
		slog.Error(ctx, "Failed to parse neighbour refresh interval %s: %v", wylisconfig.ConfigRefreshInterval, err)
		return err
	}

	slog.Debug(ctx, "Loaded %d Wylis neighbours.", len(neighbourPods))

	refreshTicker := time.NewTicker(interval)
	refreshQuit := make(chan struct{})
	go func() {
		for {
			select {
			case <-refreshTicker.C:
				err := refreshNeighbourPods(ctx)
				if err != nil {
					slog.Error(ctx, "Could not refresh neighbours, stale local pod list possible: %v", err)
				}
			case <-refreshQuit:
				refreshTicker.Stop()
				return
			}
		}
	}()

	return nil
}

func refreshNeighbourPods(ctx context.Context) error {
	neighbourPodLock.Lock()
	defer neighbourPodLock.Unlock()

	pods, err := clientSet.CoreV1().Pods(wylisconfig.ConfigNamespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", wylisconfig.ConfigIdentifierLabel, wylisconfig.ConfigIdentifierValue),
	})
	if err != nil {
		slog.Error(ctx, "Could not load initial pod list: %v", err)
		return err
	}
	if len(pods.Items) == 0 {
		// We should at least find the requesting pod itself, otherwise label selector probably misconfigured
		err = terrors.InternalService("", "Could not find any Wylis pod, did you set the identifier label and value correctly?", nil)
		slog.Error(ctx, "Error: %v", err)
		return err
	}

	neighbours := []neighbourPod{}
	for _, pod := range pods.Items {
		if pod.Status.HostIP == wylisconfig.ConfigNodeIP {
			// Don't send requests to self
			continue
		}

		if pod.Status.HostIP == "" {
			// Neighbour not ready if HostIP info not available
			continue
		}

		podIP, err := validatePodIP(extractPodIP(pod.Status))
		if err != nil {
			slog.Warn(ctx, "Found invalid pod IP: %v", err)
			continue
		}

		neighbours = append(neighbours, neighbourPod{
			podIP:  podIP,
			nodeIP: pod.Status.HostIP,
		})
	}

	if len(neighbours) == 0 {
		slog.Warn(ctx, "%s Found no neighbour pod in refresh, either we are the only node or there's a network segmentation.", wylisconfig.ConfigNodeIP)
	}

	neighbourPods = neighbours
	return nil
}

// Extract the first IP assigned to the pod, we only need to check
// each pod once, and as long as the cluster CNI is healthy, any pod
// IP assigned is guaranteed to be routable.
func extractPodIP(status corev1.PodStatus) string {
	if status.PodIP != "" {
		return status.PodIP
	}

	if len(status.PodIPs) == 0 {
		return ""
	}

	return status.PodIPs[0].IP
}

func validatePodIP(podIP string) (string, error) {
	ip := net.ParseIP(podIP)
	if ip == nil {
		return "", terrors.PreconditionFailed("invalid_pod_ip", "Invalid pod IP observed, cannot use.", map[string]string{"raw_ip": podIP})
	}

	// Because Wylis sends unmodifiable payloads, and discards response, with fixed time-outs,
	// we'll not attempt to validate whether the destination IP is within a pre-defined range here.
	// In any valid threat model this is a bad idea, but Wylis processes information from a trusted
	// apiserver only, in this case.

	return ip.String(), nil
}

func getNeighbourPods() []neighbourPod {
	neighbourPodLock.RLock()
	defer neighbourPodLock.RUnlock()

	return neighbourPods
}
