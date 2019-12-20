package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/icydoge/wylis/config"
)

var (
	incomingRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "wylis",
		Name:      "incoming_requests",
		Help:      "Successful incoming requests to this Wylis pod from others, we won't know any incoming which has failed",
	}, []string{"source_node", "target_node"})
)

var (
	outgoingRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "wylis",
		Name:      "outgoing_requests",
		Help:      "Outgoing requests from this Wylis pod to others, along with their status",
	}, []string{"source_node", "target_node", "result"})
)

var (
	outgoingTimings = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "wylis",
		Name:      "outgoing_timings",
		Help:      "Timings of successful outgoing requests from this Wylis pod to others",
		Buckets:   []float64{0, 0.01, 0.1, 0.5, 1},
	}, []string{"source_node", "target_node"})
)

func RegisterIncomingRequest(sourceIP string) {
	incomingRequests.WithLabelValues(sourceIP, config.ConfigNodeIP).Inc()
}

func RegisterOutgoingRequest(targetIP string, success bool) {
	outgoingRequests.WithLabelValues(config.ConfigNodeIP, targetIP, strconv.FormatBool(success)).Inc()
}

func RegisterOutgoingTiming(targetIP string, timing float64) {
	outgoingTimings.WithLabelValues(config.ConfigNodeIP, targetIP).Observe(timing)
}
