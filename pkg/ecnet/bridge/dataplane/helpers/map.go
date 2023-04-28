package helpers

import (
	"encoding/binary"
	"fmt"
	"net"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/miekg/dns"
)

const (
	// ECNetDNSResolveMap is the mount point of ecnet_dns_resdb map
	ECNetDNSResolveMap = "/sys/fs/bpf/ecnet_dns_resdb"

	// ECNetDNSEndpointMap is the mount point of ecnet_dns_endpt map
	ECNetDNSEndpointMap = "/sys/fs/bpf/ecnet_dns_endpt"
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
	dnsResolveMap, err = ebpf.LoadPinnedMap(ECNetDNSResolveMap, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map[%s] error: %v", ECNetDNSResolveMap, err)
	}
	dnsEndpointMap, err = ebpf.LoadPinnedMap(ECNetDNSEndpointMap, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map[%s] error: %v", err, ECNetDNSEndpointMap)
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

}

// UpdateDNSEndpointEntry updates DNS Endpoint Entry
func UpdateDNSEndpointEntry(dnsEndpointIP, dnsClusterIP string) {
	dnsEIPPtr, _ := ip2ptr(dnsEndpointIP)
	dnsCIPPtr, _ := ip2ptr(dnsClusterIP)
	err := GetDNSEndpointMap().Update(dnsEIPPtr, dnsCIPPtr, ebpf.UpdateAny)
	if err != nil {
		log.Error().Msgf("update ecnet_dns_endpt error: %v", err)
	}
}

// DeleteDNSEndpointEntry deletes DNS Endpoint Entry
func DeleteDNSEndpointEntry(dnsEndpointIP string) {
	dnsEIPPtr, _ := ip2ptr(dnsEndpointIP)
	err := GetDNSEndpointMap().Delete(dnsEIPPtr)
	if err != nil {
		log.Error().Msgf("delete ecnet_dns_endpt error: %v", err)
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
