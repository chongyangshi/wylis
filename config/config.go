package config

import (
	"fmt"
	"os"

	"github.com/monzo/terrors"
)

const SourceNodeIPHeader = "X-WYLIS-SOURCE-NODE-IP"

var (
	ConfigNodeIP             = getConfigFromOSEnv("NODE_IP", "127.0.0.1", true)
	ConfigListenAddr         = getConfigFromOSEnv("LISTEN_ADDR", "", true)
	ConfigIncomingListenPort = getConfigFromOSEnv("INCOMING_LISTEN_PORT", "9050", true)
	ConfigMetricsListenPort  = getConfigFromOSEnv("METRICS_LISTEN_PORT", "9051", true)
	ConfigKubeAPIServerAddr  = getConfigFromOSEnv("KUBE_APISERVER_ADDR", "127.0.0.1:6443", true)
	ConfigOutgoingTimeout    = getConfigFromOSEnv("KUBE_APISERVER_ADDR", "127.0.0.1:6443", true)
)

// This is intended to run inside Kubernetes as a pod of a daemonset, so we just set service Configurations from
// deployment Configuration.
func getConfigFromOSEnv(key, defaultValue string, canBeEmpty bool) string {
	envValue := os.Getenv(key)
	if envValue != "" {
		return envValue
	}

	if !canBeEmpty {
		panic(terrors.InternalService("invalid_Config", fmt.Sprintf("Config value cannot be empty: %s", key), nil))
	}

	return defaultValue
}
