package configurator

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	configv1alpha1 "github.com/flomesh-io/ErieCanal/pkg/ecnet/apis/config/v1alpha1"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/constants"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/errcode"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/trafficpolicy"
)

// The functions in this file implement the configurator.Configurator interface

// GetMeshConfig returns the MeshConfig resource corresponding to the control plane
func (c *Client) GetMeshConfig() configv1alpha1.MeshConfig {
	return c.getMeshConfig()
}

// GetECNETNamespace returns the namespace in which the ECNET controller pod resides.
func (c *Client) GetECNETNamespace() string {
	return c.ecnetNamespace
}

func marshalConfigToJSON(config configv1alpha1.MeshConfigSpec) (string, error) {
	bytes, err := json.MarshalIndent(&config, "", "    ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetMeshConfigJSON returns the MeshConfig in pretty JSON.
func (c *Client) GetMeshConfigJSON() (string, error) {
	cm, err := marshalConfigToJSON(c.getMeshConfig().Spec)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrMeshConfigMarshaling)).Msgf("Error marshaling MeshConfig %s: %+v", c.getMeshConfigCacheKey(), c.getMeshConfig())
		return "", err
	}
	return cm, nil
}

// LocalDNSProxyEnabled returns whether local DNS proxy is enabled
func (c *Client) LocalDNSProxyEnabled() bool {
	return c.getMeshConfig().Spec.Sidecar.LocalDNSProxy.Enable
}

// GetLocalDNSProxyPrimaryUpstream returns the primary upstream DNS server for local DNS Proxy
func (c *Client) GetLocalDNSProxyPrimaryUpstream() string {
	return c.getMeshConfig().Spec.Sidecar.LocalDNSProxy.PrimaryUpstreamDNSServerIPAddr
}

// GetLocalDNSProxySecondaryUpstream returns the secondary upstream DNS server for local DNS Proxy
func (c *Client) GetLocalDNSProxySecondaryUpstream() string {
	return c.getMeshConfig().Spec.Sidecar.LocalDNSProxy.SecondaryUpstreamDNSServerIPAddr
}

// GetSidecarLogLevel returns the sidecar log level
func (c *Client) GetSidecarLogLevel() string {
	logLevel := c.getMeshConfig().Spec.Sidecar.LogLevel
	if logLevel != "" {
		return logLevel
	}
	return constants.DefaultSidecarLogLevel
}

// GetSidecarClass returns the sidecar class
func (c *Client) GetSidecarClass() string {
	class := c.getMeshConfig().Spec.Sidecar.SidecarClass
	if class == "" {
		class = constants.SidecarClassPipy
	}
	return class
}

// GetSidecarImage returns the sidecar image
func (c *Client) GetSidecarImage() string {
	image := c.getMeshConfig().Spec.Sidecar.SidecarImage
	if len(image) == 0 {
		image = os.Getenv("ECNET_DEFAULT_SIDECAR_IMAGE")
	}
	return image
}

// GetProxyServerPort returns the port on which the Discovery Service listens for new connections from Sidecars
func (c *Client) GetProxyServerPort() uint32 {
	return c.getMeshConfig().Spec.Sidecar.ProxyServerPort
}

// GetRepoServerIPAddr returns the ip address of RepoServer
func (c *Client) GetRepoServerIPAddr() string {
	ipAddr := os.Getenv("ECNET_REPO_SERVER_IPADDR")
	if len(ipAddr) == 0 {
		ipAddr = c.getMeshConfig().Spec.RepoServer.IPAddr
	}
	if len(ipAddr) == 0 {
		ipAddr = "127.0.0.1"
	}
	return ipAddr
}

// GetRepoServerCodebase returns the codebase of RepoServer
func (c *Client) GetRepoServerCodebase() string {
	codebase := os.Getenv("ECNET_REPO_SERVER_CODEBASE")
	if len(codebase) == 0 {
		codebase = c.getMeshConfig().Spec.RepoServer.Codebase
	}
	if len(codebase) > 0 && strings.HasSuffix(codebase, "/") {
		codebase = strings.TrimSuffix(codebase, "/")
	}
	if len(codebase) > 0 && strings.HasPrefix(codebase, "/") {
		codebase = strings.TrimPrefix(codebase, "/")
	}
	return codebase
}

// GetConfigResyncInterval returns the duration for resync interval.
// If error or non-parsable value, returns 0 duration
func (c *Client) GetConfigResyncInterval() time.Duration {
	resyncDuration := c.getMeshConfig().Spec.Sidecar.ConfigResyncInterval
	duration, err := time.ParseDuration(resyncDuration)
	if err != nil {
		log.Warn().Msgf("Error parsing config resync interval: %s", duration)
		return time.Duration(0)
	}
	return duration
}

// GetGlobalPluginChains returns plugin chains
func (c *Client) GetGlobalPluginChains() map[string][]trafficpolicy.Plugin {
	pluginChainMap := make(map[string][]trafficpolicy.Plugin)
	pluginChainSpec := c.getMeshConfig().Spec.PluginChains

	inboundTCPChains := make([]trafficpolicy.Plugin, 0)
	for _, plugin := range pluginChainSpec.InboundTCPChains {
		if plugin.Disable {
			continue
		}
		inboundTCPChains = append(inboundTCPChains, trafficpolicy.Plugin{
			Name:     plugin.Plugin,
			Priority: plugin.Priority,
			BuildIn:  true,
		})
	}

	inboundHTTPChains := make([]trafficpolicy.Plugin, 0)
	for _, plugin := range pluginChainSpec.InboundHTTPChains {
		if plugin.Disable {
			continue
		}
		inboundHTTPChains = append(inboundHTTPChains, trafficpolicy.Plugin{
			Name:     plugin.Plugin,
			Priority: plugin.Priority,
			BuildIn:  true,
		})
	}

	outboundTCPChains := make([]trafficpolicy.Plugin, 0)
	for _, plugin := range pluginChainSpec.OutboundTCPChains {
		if plugin.Disable {
			continue
		}
		outboundTCPChains = append(outboundTCPChains, trafficpolicy.Plugin{
			Name:     plugin.Plugin,
			Priority: plugin.Priority,
			BuildIn:  true,
		})
	}

	outboundHTTPChains := make([]trafficpolicy.Plugin, 0)
	for _, plugin := range pluginChainSpec.OutboundHTTPChains {
		if plugin.Disable {
			continue
		}
		outboundHTTPChains = append(outboundHTTPChains, trafficpolicy.Plugin{
			Name:     plugin.Plugin,
			Priority: plugin.Priority,
			BuildIn:  true,
		})
	}

	pluginChainMap["inbound-tcp"] = inboundTCPChains
	pluginChainMap["inbound-http"] = inboundHTTPChains
	pluginChainMap["outbound-tcp"] = outboundTCPChains
	pluginChainMap["outbound-http"] = outboundHTTPChains
	return pluginChainMap
}