// Package service models an instance of a service managed by ECNET controller and utility routines associated with it.
package service

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

// MeshService is the struct representing a service (Kubernetes or otherwise) within the service mesh.
type MeshService struct {
	// If the service resides on a Kubernetes service, this would be the Kubernetes namespace.
	Namespace string

	// The name of the service. May include instance (e.g. pod) information if the backing service
	// doesn't have a single, stable ip address. For example, a MeshService created by a headless
	// Kubernetes service named mysql-headless, will have the name "mysql.mysql-headless"
	// A MeshService created by a normal ClusterIP service named mysql will be named "mysql"
	// This imposes a restriction that service names cannot contain "." (which is already
	// the case in kubernetes). Thus, MeshService.Name will be of the form: [subdomain.]providerKey
	Name string

	// Port is the port number that clients use to access the service.
	// This can be different than MeshService.TargetPort which represents the actual port number
	// the application is accepting connections on.
	// Port maps to ServicePort.Port in k8s: https://pkg.go.dev/k8s.io/api/core/v1#ServicePort
	Port uint16

	// TargetPort is the port number on which an application accept traffic directed to this MeshService
	// This can be different than MeshService.Port in k8s.
	// TargetPort maps to ServicePort.TargetPort in k8s: https://pkg.go.dev/k8s.io/api/core/v1#ServicePort
	TargetPort uint16

	// Protocol is the protocol served by the service's port
	Protocol string

	// ServiceImportUID is the uid of service import
	ServiceImportUID types.UID
}

// NamespacedKey is the key (i.e. namespace + ProviderKey()) with which to lookup the backing service within the provider
func (ms MeshService) NamespacedKey() string {
	return fmt.Sprintf("%s/%s", ms.Namespace, ms.ProviderKey())
}

// Subdomain is an optional subdomain for this MeshService
// TODO: possibly memoize if performance suffers
func (ms *MeshService) Subdomain() string {
	nameComponents := strings.Split(ms.Name, ".")
	if len(nameComponents) == 1 {
		return ""
	}
	return nameComponents[0]
}

// ProviderKey represents the name of the original entity from which this MeshService was created (e.g. a Kubernetes service name)
// TODO: possibly memoize if performance suffers
func (ms *MeshService) ProviderKey() string {
	nameComponents := strings.Split(ms.Name, ".")

	return nameComponents[len(nameComponents)-1]
}

// IsMultiClusterService checks whether it is a multi cluster service
func (ms *MeshService) IsMultiClusterService() bool {
	return len(ms.ServiceImportUID) > 0
}

// SiblingTo returns true if svc and ms are derived from the same resource
// in the service provder (based on namespace and provider key)
func (ms MeshService) SiblingTo(svc MeshService) bool {
	return ms.NamespacedKey() == svc.NamespacedKey()
}

// String returns the string representation of the given MeshService.
// SHOULD NOT BE USED AS A MAPPING FOR ANYTHING. Use NamespacedKey and Subdomain
func (ms MeshService) String() string {
	return fmt.Sprintf("%s/%s", ms.Namespace, ms.Name)
}

// SidecarClusterName is the name of the cluster corresponding to the MeshService in Sidecar
func (ms MeshService) SidecarClusterName() string {
	return fmt.Sprintf("%s/%s|%d", ms.Namespace, ms.Name, ms.TargetPort)
}

// FQDN is similar to String(), but uses a dot separator and is in a different order.
func (ms MeshService) FQDN() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", ms.Name, ms.Namespace)
}

// OutboundTrafficMatchName returns the MeshService outbound traffic match name
func (ms MeshService) OutboundTrafficMatchName() string {
	return fmt.Sprintf("outbound_%s_%d_%s", ms, ms.Port, ms.Protocol)
}

// ClusterName is a type for a service name
type ClusterName string

// String returns the given ClusterName type as a string
func (c ClusterName) String() string {
	return string(c)
}

// WeightedCluster is a struct of a cluster and is weight that is backing a service
type WeightedCluster struct {
	ClusterName ClusterName `json:"cluster_name:omitempty"`
	Weight      int         `json:"weight:omitempty"`
}

// Provider is an interface to be implemented by components abstracting Kubernetes, and other compute/cluster providers
type Provider interface {

	// ListServices returns a list of services that are part of monitored namespaces
	ListServices() []MeshService

	// GetID returns the unique identifier of the Provider
	GetID() string
}
