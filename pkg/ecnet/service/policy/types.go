// Package policy defines the types to represent traffic policies internally in the ECNET control plane, and
// utility routines to process them.
package policy

import (
	mapset "github.com/deckarep/golang-set"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service"
)

// TrafficSpecName is the namespaced name of the SMI TrafficSpec
type TrafficSpecName string

// TrafficSpecMatchName is the  name of a match in SMI TrafficSpec
type TrafficSpecMatchName string

// PathMatchType is the type used to represent the patch matching type: regex, exact, or prefix
type PathMatchType int

const (
	// PathMatchRegex is the type used to specify regex based path matching
	PathMatchRegex PathMatchType = iota

	// PathMatchExact is the type used to specify exact path matching
	PathMatchExact PathMatchType = iota

	// PathMatchPrefix is the type used to specify prefix based path matching
	PathMatchPrefix PathMatchType = iota
)

// HTTPRouteMatch is a struct to represent an HTTP route match comprised of an HTTP path, path matching type, methods, and headers
type HTTPRouteMatch struct {
	Path          string            `json:"path:omitempty"`
	PathMatchType PathMatchType     `json:"path_match_type:omitempty"`
	Methods       []string          `json:"methods:omitempty"`
	Headers       map[string]string `json:"headers:omitempty"`
}

// HTTPRouteMatchWithWeightedClusters is a struct to represent an HTTP route match comprised of WeightedClusters, HTTPRouteMatches
type HTTPRouteMatchWithWeightedClusters struct {
	UpstreamClusters []service.WeightedCluster
	RouteMatches     []HTTPRouteMatch
	HasSplitMatches  bool
}

// TCPRouteMatch is a struct to represent a TCP route matching based on ports
type TCPRouteMatch struct {
	Ports []uint16 `json:"ports:omitempty"`
}

// RouteWeightedClusters is a struct of an HTTPRoute, associated weighted clusters and the domains
type RouteWeightedClusters struct {
	HTTPRouteMatch   HTTPRouteMatch `json:"http_route_match:omitempty"`
	WeightedClusters mapset.Set     `json:"weighted_clusters:omitempty"`
}

// Rule is a struct that represents which authenticated principals can access a Route.
// A principal is of the form <service-identity>.<trust-domain>. It can also contain wildcards.
type Rule struct {
	Route RouteWeightedClusters `json:"route:omitempty"`
	// Principals contain the trust domain already while identities do not.
	AllowedPrincipals mapset.Set `json:"allowed_principals:omitempty"`
}

// OutboundTrafficPolicy is a struct that associates a list of Routes with outbound traffic on a set of Hostnames
type OutboundTrafficPolicy struct {
	Name      string                   `json:"name:omitempty"`
	Hostnames []string                 `json:"hostnames"`
	Routes    []*RouteWeightedClusters `json:"routes:omitempty"`
}

// OutboundMeshTrafficPolicy is the type used to represent the outbound mesh traffic policy configurations
// applicable to a downstream client.
type OutboundMeshTrafficPolicy struct {
	// TrafficMatches defines the list of traffic matches for matching outbound mesh traffic.
	// The matches specified are used to match outbound traffic as mesh traffic, and
	// subject matching traffic to mesh traffic policies.
	TrafficMatches []*TrafficMatch

	// HTTPRouteConfigsPerPort defines the outbound mesh HTTP route configurations per port.
	// Mesh HTTP routes are grouped based on their port to avoid route conflicts that
	// can arise when the same host headers are to be routed differently based on the port.
	HTTPRouteConfigsPerPort map[int][]*OutboundTrafficPolicy

	// ClustersConfigs defines the list of mesh cluster configurations.
	// The specified config is used to program clusters corresponding to
	// mesh destinations.
	ClustersConfigs []*MeshClusterConfig

	// ServicesResolvableSet defines the dns database
	ServicesResolvableSet map[string][]interface{}
}

// MeshClusterConfig is the type used to represent a cluster configuration that is programmed
// for either:
// 1. A downstream to connect to an upstream cluster, OR
// 2. An upstream cluster to accept traffic from a downstream
type MeshClusterConfig struct {
	// Name is the cluster's name, as referenced in an RDS route or TCP proxy filter
	Name string

	// Service is the MeshService the cluster corresponds to.
	Service service.MeshService

	// Address is the IP address/hostname of this cluster
	// This is set for local (upstream) clusters accepting traffic from a downstream client.
	// +optional
	Address string

	// Port is the port on which this cluster is listening for downstream connections.
	// This is set for local (upstream) clusters accepting traffic from a downstream client.
	// +optional
	Port uint32
}

// TrafficMatch is the type used to represent attributes used to match traffic
type TrafficMatch struct {
	// DestinationPort defines the destination port number
	DestinationPort int

	// DestinationProtocol defines the protocol served by DestinationPort
	DestinationProtocol string

	// ServerNames defines the list of server names to be used as SNI when the
	// DestinationProtocol is TLS based, ex. when the DestinationProtocol is 'https'
	// +optional
	ServerNames []string

	// Cluster defines the cluster associated with this TrafficMatch, if possible.
	// A TrafficMatch corresponding to an HTTP based cluster will not make use of
	// this property since the cluster is determined based on the computed routes.
	// A TraficMatch corresponding to a TCP based cluster will make use of this
	// property to associate the match with the corresponding cluster.
	// +optional
	Cluster string

	// Name for the match object
	// +optional
	Name string

	// WeightedClusters is list of weighted clusters that this match should
	// route traffic to. This is used by TCP based mesh clusters.
	// +optional
	WeightedClusters []service.WeightedCluster
}

// Plugin defines plugin
type Plugin struct {
	// Name defines the Name of the plugin.
	Name string

	// priority defines the priority of the plugin.
	Priority float32

	// Script defines pipy script used by the PlugIn.
	Script string

	// BuildIn indicates PlugIn type.
	BuildIn bool
}
