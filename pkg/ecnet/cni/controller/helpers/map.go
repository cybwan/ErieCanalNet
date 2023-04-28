// Package helpers implements ebpf helpers.
package helpers

import (
	"encoding/binary"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/cni/config"
	"github.com/miekg/dns"
	"net"
	"unsafe"
)

var (
	dnsResolveMap  *ebpf.Map
	dnsEndpointMap *ebpf.Map
)

const (
	MaxDnsNameLength = 256
)

type DNSQuery struct {
	QType  uint16
	QClass uint16
	Name   [MaxDnsNameLength]byte
}

type AAAARecord struct {
	Addr [4]byte
	Ttl  uint32
}

// InitLoadPinnedMap init, load and pinned maps
func InitLoadPinnedMap() error {
	var err error
	dnsResolveMap, err = ebpf.LoadPinnedMap(config.ECNetDNSResolveMap, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map[%s] error: %v", config.ECNetDNSResolveMap, err)
	}
	dnsEndpointMap, err = ebpf.LoadPinnedMap(config.ECNetDNSEndpointMap, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map[%s] error: %v", err, config.ECNetDNSEndpointMap)
	}
	return nil
}

// GetDNSResolveMap returns dns resolve map
func GetDNSResolveMap() *ebpf.Map {
	if dnsResolveMap == nil {
		_ = InitLoadPinnedMap()
	}
	return dnsResolveMap
}

// GetDNSEndpointMap returns dns endpoint map
func GetDNSEndpointMap() *ebpf.Map {
	if dnsEndpointMap == nil {
		_ = InitLoadPinnedMap()
	}
	return dnsEndpointMap
}

func DoTest() {
	ttl := uint32(60)
	names := []string{
		"pipy-ok.pipy.",
		"pipy-ok.pipy.svc.",
		"pipy-ok.pipy.svc.cluster.",
		"pipy-ok.pipy.svc.cluster.local.",
		"pipy-ok.pipy.cluster.",
		"pipy-ok.pipy.cluster.local.",
	}
	qTypes := []uint16{dns.TypeA, dns.TypeAAAA}
	for _, name := range names {
		for _, qType := range qTypes {
			k := DNSQuery{
				QType:  qType,
				QClass: dns.ClassINET,
			}
			dns.PackDomainName(name, k.Name[:], 0, nil, false)
			v := AAAARecord{
				Ttl: ttl,
			}
			binary.BigEndian.PutUint32(v.Addr[:], bridgeIPInt)
			err := GetDNSResolveMap().Update(&k, &v, ebpf.UpdateAny)
			if err != nil {
				log.Error().Msgf("update ecnet_dns_resdb error: %v", err)
			}
		}
	}

	dnsEIp1 := "10.244.0.2"
	dnsEIp2 := "10.244.0.3"
	dnsCCp := "10.96.0.10"

	dnsEIpPtr1, _ := ip2ptr(dnsEIp1)
	dnsEIpPtr2, _ := ip2ptr(dnsEIp2)
	dnsCCpPtr, _ := ip2ptr(dnsCCp)

	err := GetDNSEndpointMap().Update(dnsEIpPtr1, dnsCCpPtr, ebpf.UpdateAny)
	if err != nil {
		log.Error().Msgf("update ecnet_dns_endpt error: %v", err)
	}
	err = GetDNSEndpointMap().Update(dnsEIpPtr2, dnsCCpPtr, ebpf.UpdateAny)
	if err != nil {
		log.Error().Msgf("update ecnet_dns_endpt error: %v", err)
	}
}

func ip2ptr(ipstr string) (unsafe.Pointer, error) {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return nil, fmt.Errorf("error parse ip: %s", ipstr)
	}
	//#nosec G103
	return unsafe.Pointer(&ip.To4()[0]), nil
}
