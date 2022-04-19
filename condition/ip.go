package condition

import (
	"net"

	"github.com/brexhq/substation/internal/json"
)

// IP implements the Inspector interface for evaluating IP address data. More information is available in the README.
type IP struct {
	Key      string `mapstructure:"key"`
	Function string `mapstructure:"function"`
	Negate   bool   `mapstructure:"negate"`
}

// Inspect evaluates the type and usage of an IP address.
func (c IP) Inspect(data []byte) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(data)
	} else {
		check = json.Get(data, c.Key).String()
	}

	ip := net.ParseIP(check)

	var matched bool
	switch f := c.Function; f {
	case "loopback":
		matched = ip.IsLoopback()
	case "multicast":
		matched = ip.IsMulticast()
	case "multicast_link_local":
		matched = ip.IsLinkLocalMulticast()
	case "private":
		matched = ip.IsPrivate()
	case "unicast_global":
		matched = ip.IsGlobalUnicast()
	case "unicast_link_local":
		matched = ip.IsLinkLocalUnicast()
	case "unspecified":
		matched = ip.IsUnspecified()
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
