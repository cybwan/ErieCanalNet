package helpers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

var (
	bridgeIPInt  uint32
	bridgeIPAddr net.IP
)

// GetBridgeIP retrieves cni bridge veth's ipv4 addr
func GetBridgeIP(bridgeEth string) (ipAddr net.IP, ipInt uint32) {
	var err error
	for {
		ipAddr, ipInt, err = waitBridgeIP(bridgeEth)
		if err == nil && ipInt > 0 {
			break
		}
		if err != nil {
			log.Warn().Msgf("fail retrieving cni bridge veth[%s]'s ipv4 addr:%v, and retring...", bridgeEth, err)
			time.Sleep(time.Second * 5)
		}
	}
	return
}

func waitBridgeIP(bridgeEth string) (net.IP, uint32, error) {
	if bridgeIPInt == 0 {
		found := false
		if ifaces, err := net.Interfaces(); err == nil {
			for _, iface := range ifaces {
				if iface.Flags&net.FlagUp != 0 && strings.HasPrefix(iface.Name, bridgeEth) {
					if addrs, addrErr := iface.Addrs(); addrErr == nil {
						for _, addr := range addrs {
							if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
								if bridgeIPAddr = ipnet.IP.To4(); bridgeIPAddr != nil {
									bridgeIPInt = binary.BigEndian.Uint32(bridgeIPAddr)
									found = true
									break
								}
							}
						}
					} else {
						return bridgeIPAddr, bridgeIPInt, fmt.Errorf("unexpected exit err: %v", err)
					}
					break
				}
			}
		} else {
			return bridgeIPAddr, bridgeIPInt, fmt.Errorf("unexpected exit err: %v", err)
		}
		if !found {
			return bridgeIPAddr, bridgeIPInt, fmt.Errorf("unexpected retrieves cni bridge veth[%s]'s ipv4 addr", bridgeEth)
		}
	}
	return bridgeIPAddr, bridgeIPInt, nil
}
