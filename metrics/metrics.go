package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/chongyangshi/wylis/config"
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

var (
	outgoingStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "wylis",
		Name:      "outgoing_status",
		Help:      "Current proportion of successful outgoing requests from this Wylis pod.",
	}, []string{"source_node"})
)

func RegisterIncomingRequest(sourceNode string) {
	incomingRequests.WithLabelValues(sourceNode, config.ConfigNodeIP).Inc()
}

func RegisterOutgoingRequest(targetNode string, success bool) {
	outgoingRequests.WithLabelValues(config.ConfigNodeIP, targetNode, strconv.FormatBool(success)).Inc()
}

func RegisterOutgoingTiming(targetNode string, timing float64) {
	outgoingTimings.WithLabelValues(config.ConfigNodeIP, targetNode).Observe(timing)
}

func RegisterOutgoingStatus(numSuccess, total int) {
	if numSuccess > total {
		numSuccess = total // Invalid input, cap at 1.
	}

	successRate := float64(numSuccess) / float64(total)
	outgoingStatus.WithLabelValues(config.ConfigNodeIP).Set(successRate)
}
