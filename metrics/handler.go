package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/icydoge/wylis/config"
)

func Init() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf("%s:%s", config.ConfigListenAddr, config.ConfigMetricsListenPort), nil)
}
