package outgoing

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Normally, when calling pods of a Kubernetes service without using a separate
// service proxy, the client should call the cluster IP for kube-proxy to forward
// the reuqest to an arbitrary destination pod which could serve the it. We can't
// do that but instead must have a full picture of pods in the Wylis daemonset, as
// we need to reach pods in the daemonset running on all nodes.

func initClusterClient(ctx context.Context) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

type neighbour struct {
	targetIP string
}
