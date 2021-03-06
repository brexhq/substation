package condition

import (
	"fmt"
	"net"

	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// IPInvalidType is returned when the IP inspector is configured with an invalid type.
const IPInvalidType = errors.Error("IPInvalidType")

/*
IP evaluates IP addresses by their type and usage. This inspector uses the standard library's net package to identify the type and usage of the address (more information is available here: https://pkg.go.dev/net#IP).

The inspector has these settings:
	Key (optional):
		the JSON key-value to retrieve for inspection
	Type:
		the IP address type used during inspection
		must be one of:
			loopback: valid loopback address
			multicast: valid multicast address
			multicast_link_local: valid link local multicast address
			private: valid private address
			unicast_global: valid global unicast address
			unicast_link_local: valid link local unicast address
			unspecified: valid "unspecified" address (e.g., 0.0.0.0, ::)
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	json:
		{"ip_address":"10.0.0.1"} == private
	data:
		10.0.0.1 == private

The inspector uses this Jsonnet configuration:
	{
		type: 'ip',
		settings: {
			key: 'ip_address',
			type: 'private',
		},
	}
*/
type IP struct {
	Key    string `json:"key"`
	Type   string `json:"type"`
	Negate bool   `json:"negate"`
}

// Inspect evaluates data with the IP inspector.
func (c IP) Inspect(data []byte) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(data)
	} else {
		check = json.Get(data, c.Key).String()
	}

	ip := net.ParseIP(check)

	var matched bool
	switch s := c.Type; s {
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
	default:
		return false, fmt.Errorf("inspector settings %v: %v", c, IPInvalidType)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
