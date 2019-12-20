package config

import (
	"fmt"
	"os"

	"github.com/monzo/terrors"
)

const SourceNodeIPHeader = "X-WYLIS-SOURCE-NODE-IP"

var (
	ConfigNamespace          = getConfigFromOSEnv("NAMESPACE", "default", true)
	ConfigIdentifierLabel    = getConfigFromOSEnv("IDENTIFIER_LABEL", "app", true)
	ConfigIdentifierValue    = getConfigFromOSEnv("IDENTIFIER_VALUE", "wylis", true)
	ConfigNodeIP             = getConfigFromOSEnv("NODE_IP", "127.0.0.1", true)
	ConfigListenAddr         = getConfigFromOSEnv("LISTEN_ADDR", "", true)
	ConfigIncomingListenPort = getConfigFromOSEnv("INCOMING_LISTEN_PORT", "9050", true)
	ConfigMetricsListenPort  = getConfigFromOSEnv("METRICS_LISTEN_PORT", "9051", true)
	ConfigOutgoingTimeout    = getConfigFromOSEnv("OUTGOING_TIMEOUT", "5s", true)
	ConfigOutgoingInterval   = getConfigFromOSEnv("OUTGOING_INTERVAL", "10s", true)
	ConfigRefreshInterval    = getConfigFromOSEnv("REFRESH_INTERVAL", "30s", true)
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
