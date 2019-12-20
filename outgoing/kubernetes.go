package outgoing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	wylisconfig "github.com/icydoge/wylis/config"
)

// Normally, when calling pods of a Kubernetes service without using a separate
// service proxy, the client should call the cluster IP for kube-proxy to forward
// the reuqest to an arbitrary destination pod which could serve the it. We can't
// do that but instead must have a full picture of pods in the Wylis daemonset, as
// we need to reach pods in the daemonset running on all nodes.
var (
	clientSet        = &kubernetes.Clientset{}
	neighbourPodLock = sync.RWMutex{}
	neighbourPodIPs  = []string{}
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

	slog.Debug(ctx, "Loaded %d Wylis neighbours.", len(neighbourPodIPs))

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
		err = terrors.InternalService("", "Could not find any Wylis pod, did you set the identifier label and value correctly?", nil)
		slog.Error(ctx, "Error: %v", err)
		return err
	}

	neighbours := []string{}
	for _, pod := range pods.Items {
		if pod.Status.HostIP == wylisconfig.ConfigNodeIP {
			// Don't send requests to self
			continue
		}

		neighbours = append(neighbours, pod.Status.PodIP)
	}

	neighbourPodIPs = neighbours
	return nil
}

func getNeighbourPods() []string {
	neighbourPodLock.RLock()
	defer neighbourPodLock.RUnlock()

	return neighbourPodIPs
}
