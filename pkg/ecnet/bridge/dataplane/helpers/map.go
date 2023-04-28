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
	maxDNSNameLength = 256
)

type dnsQuery struct {
	QType  uint16
	QClass uint16
	Name   [maxDNSNameLength]byte
}

type aRecord struct {
	Addr [4]byte
	TTL  uint32
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

// UpdateDNSResolveEntry updates DNS Resolve Entry
func UpdateDNSResolveEntry(name string, ttl uint32) {
	qTypes := []uint16{dns.TypeA, dns.TypeAAAA}
	for _, qType := range qTypes {
		k := dnsQuery{
			QType:  qType,
			QClass: dns.ClassINET,
		}
		_, err := dns.PackDomainName(fmt.Sprintf("%s.", name), k.Name[:], 0, nil, false)
		if err != nil {
			log.Error().Msgf("update ecnet_dns_resdb entry error: %v", err)
			continue
		}
		v := aRecord{
			TTL: ttl,
		}
		binary.BigEndian.PutUint32(v.Addr[:], bridgeIPInt)
		err = GetDNSResolveMap().Update(&k, &v, ebpf.UpdateAny)
		if err != nil {
			log.Error().Msgf("update ecnet_dns_resdb entry error: %v", err)
		}
	}
}

// DeleteDNSResolveEntry deletes DNS Resolve Entry
func DeleteDNSResolveEntry(name string) {
	qTypes := []uint16{dns.TypeA, dns.TypeAAAA}
	for _, qType := range qTypes {
		k := dnsQuery{
			QType:  qType,
			QClass: dns.ClassINET,
		}
		_, err := dns.PackDomainName(name, k.Name[:], 0, nil, false)
		if err != nil {
			log.Error().Msgf("update ecnet_dns_resdb entry error: %v", err)
			continue
		}
		err = GetDNSResolveMap().Delete(&k)
		if err != nil {
			log.Error().Msgf("delete ecnet_dns_resdb entry error: %v", err)
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
