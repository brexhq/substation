package condition

import (
	"net"

	iconfig "github.com/brexhq/substation/internal/config"
)

type networkIPConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *networkIPConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func netIPIsLoopback(ip net.IP) bool {
	return ip.IsLoopback()
}

func netIPIsPrivate(ip net.IP) bool {
	return ip.IsPrivate()
}

func netIPIsMulticast(ip net.IP) bool {
	return ip.IsMulticast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast()
}

func netIPIsUnicast(ip net.IP) bool {
	return ip.IsGlobalUnicast() || ip.IsLinkLocalUnicast()
}

func netIPIsUnspecified(ip net.IP) bool {
	return ip.IsUnspecified()
}

func netIPIsValid(ip net.IP) bool {
	return ip != nil
}

func netIPIsPublic(ip net.IP) bool {
	return netIPIsValid(ip) && !netIPIsLoopback(ip) && !netIPIsPrivate(ip) && !netIPIsMulticast(ip) && !netIPIsUnicast(ip) && !netIPIsUnspecified(ip)
}
