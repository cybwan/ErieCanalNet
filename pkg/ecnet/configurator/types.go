// Package configurator implements the Configurator interface that provides APIs to retrieve ECNET control plane configurations.
package configurator

import (
	"time"

	configv1alpha1 "github.com/flomesh-io/ErieCanal/pkg/ecnet/apis/config/v1alpha1"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/k8s/informers"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/logger"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service/policy"
)

var (
	log = logger.New("configurator")
)

// Client is the type used to represent the Kubernetes Client for the config.flomesh.io API group
type Client struct {
	ecnetNamespace  string
	informers       *informers.InformerCollection
	ecnetConfigName string
}

// Configurator is the ctrlplane interface for K8s namespaces
type Configurator interface {
	// GetEcnetConfig returns the EcnetConfig resource corresponding to the control plane
	GetEcnetConfig() configv1alpha1.EcnetConfig

	// GetEcnetNamespace returns the namespace in which ECNET ctrlplane pod resides
	GetEcnetNamespace() string

	// GetEcnetConfigJSON returns the EcnetConfig in pretty JSON (human readable)
	GetEcnetConfigJSON() (string, error)

	// GetSidecarLogLevel returns the sidecar log level
	GetSidecarLogLevel() string

	// GetProxyServerPort returns the port on which the Discovery Service listens for new connections from Sidecars
	GetProxyServerPort() uint32

	// GetRepoServerIPAddr returns the ip address of RepoServer
	GetRepoServerIPAddr() string

	// GetRepoServerCodebase returns the codebase of RepoServer
	GetRepoServerCodebase() string

	// GetConfigResyncInterval returns the duration for resync interval.
	// If error or non-parsable value, returns 0 duration
	GetConfigResyncInterval() time.Duration

	// GetGlobalPluginChains returns plugin chains
	GetGlobalPluginChains() map[string][]policy.Plugin
}
