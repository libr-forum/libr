package addressutils

import (
	"net"
	"strconv"

	"github.com/pion/stun"
)

func isPrivateIPv4(ip net.IP) bool {
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range privateBlocks {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func GetPrivateIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "UNKNOWN"
	}
	for _, address := range addrs {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				if isPrivateIPv4(ip4) {
					return ip4.String()
				}
			}
		}
	}
	return "UNKNOWN"
}

func GetPublicIP() string {
	c, err := stun.Dial("udp4", "stun.l.google.com:19302")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	var xorAddr stun.XORMappedAddress
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	if err := c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			panic(res.Error)
		}
		if err := xorAddr.GetFrom(res.Message); err != nil {
			panic(err)
		}
	}); err != nil {
		panic(err)
	}

	if xorAddr.IP.To4() == nil {
		panic("STUN returned an IPv6 address; IPv4 not available")
	}

	peerAd := xorAddr.IP.String() + ":" + strconv.Itoa(xorAddr.Port)
	return peerAd
}
