package server

import (
	"net"
	"strings"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/bridge/dataplane/helpers"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/catalog"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service"
)

// BridgeNodeWatcherJob is the job to generate pipy policy json
type BridgeNodeWatcherJob struct {
	bridgeServer *Server

	// Optional waiter
	done chan struct{}
}

// GetDoneCh returns the channel, which when closed, indicates the job has been finished.
func (job *BridgeNodeWatcherJob) GetDoneCh() <-chan struct{} {
	return job.done
}

// Run is the logic unit of job
func (job *BridgeNodeWatcherJob) Run() {
	defer close(job.done)

	s := job.bridgeServer
	syncDNSEndpoints(s.catalog, s.dnsEndpoints)
	syncDNSResolves(s.catalog, s.dnsResolves)
}

// JobName implementation for this job, for logging purposes
func (job *BridgeNodeWatcherJob) JobName() string {
	return "bridgeJob"
}

func syncDNSEndpoints(mc catalog.MeshCataloger, dnsEndpointsMap map[string]string) {
	dnsSvc := service.MeshService{
		Namespace: "kube-system",
		Name:      "kube-dns",
		Port:      53,
		Protocol:  "udp",
	}
	kubeController := mc.GetKubeController()
	svc := kubeController.GetService(dnsSvc)
	if dnsEndpoints, err := kubeController.GetEndpoints(dnsSvc); err == nil {
		latestEndpointsMap := make(map[string]string)
		for _, dnsEndpoint := range dnsEndpoints.Subsets {
			for _, port := range dnsEndpoint.Ports {
				if port.Port != int32(dnsSvc.Port) {
					continue
				}
				for _, address := range dnsEndpoint.Addresses {
					if dnsSvc.Subdomain() != "" && dnsSvc.Subdomain() != address.Hostname {
						// if there's a subdomain on this meshservice, make sure it matches the endpoint's hostname
						continue
					}
					ip := net.ParseIP(address.IP)
					if ip == nil {
						log.Error().Msgf("Error parsing endpoint IP address %s for MeshService %s", address.IP, dnsSvc)
						continue
					}
					latestEndpointsMap[address.IP] = svc.Spec.ClusterIP
				}
			}
		}
		for eip, cip := range latestEndpointsMap {
			if pcip, exist := dnsEndpointsMap[eip]; !exist {
				helpers.UpdateDNSEndpointEntry(eip, cip)
			} else if !strings.EqualFold(cip, pcip) {
				dnsEndpointsMap[eip] = cip
				helpers.UpdateDNSEndpointEntry(eip, cip)
			}
		}
		for eip := range dnsEndpointsMap {
			if _, exist := latestEndpointsMap[eip]; !exist {
				delete(dnsEndpointsMap, eip)
				helpers.DeleteDNSEndpointEntry(eip)
			}
		}
	} else {
		log.Error().Err(err)
	}
}

func syncDNSResolves(mc catalog.MeshCataloger, dnsResolvesMap map[string]uint8) {
	outboundTrafficPolicy := mc.GetOutboundMeshTrafficPolicy()
	if len(outboundTrafficPolicy.HTTPRouteConfigsPerPort) == 0 {
		return
	}
	latestResolvesMap := make(map[string]uint8)
	for _, outboundTrafficPolicies := range outboundTrafficPolicy.HTTPRouteConfigsPerPort {
		if len(outboundTrafficPolicies) == 0 {
			continue
		}
		for _, outboundPolicy := range outboundTrafficPolicies {
			if len(outboundPolicy.Hostnames) == 0 {
				continue
			}
			for _, hostname := range outboundPolicy.Hostnames {
				if strings.Contains(hostname, ":") {
					continue
				}
				latestResolvesMap[hostname] = 0
			}
		}
	}
	for hostname := range latestResolvesMap {
		if _, exist := dnsResolvesMap[hostname]; !exist {
			helpers.UpdateDNSResolveEntry(hostname, 60)
		}
	}
	for hostname := range dnsResolvesMap {
		if _, exist := latestResolvesMap[hostname]; !exist {
			delete(dnsResolvesMap, hostname)
			helpers.DeleteDNSResolveEntry(hostname)
		}
	}
}
