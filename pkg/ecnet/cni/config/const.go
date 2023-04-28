// Package config defines the constants that are used by multiple other packages within ECNET.
package config

const (
	// ECNetDNSResolveMap is the mount point of ecnet_dns_resdb map
	ECNetDNSResolveMap = "/sys/fs/bpf/ecnet_dns_resdb"

	// ECNetDNSEndpointMap is the mount point of ecnet_dns_endpt map
	ECNetDNSEndpointMap = "/sys/fs/bpf/ecnet_dns_endpt"
)
