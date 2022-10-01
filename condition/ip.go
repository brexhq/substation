package condition

import (
	"context"
	"fmt"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// ipInvalidType is returned when the IP inspector is configured with an invalid type.
const ipInvalidType = errors.Error("ipInvalidType")

/*
IP evaluates IP addresses by their type and usage. This inspector uses the standard library's net package to identify the type and usage of the address (more information is available here: https://pkg.go.dev/net#IP).

The inspector has these settings:
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
	Key (optional):
		the JSON key-value to retrieve for inspection
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	JSON:
		{"ip_address":"10.0.0.1"} == private

	data:
		10.0.0.1 == private

When loaded with a factory, the inspector uses this JSON configuration:
	{
		"type": "ip",
		"settings": {
			"type": "private"
		}
	}
*/
type IP struct {
	Type   string `json:"type"`
	Key    string `json:"key"`
	Negate bool   `json:"negate"`
}

// Inspect evaluates encapsulated data with the IP inspector.
func (c IP) Inspect(ctx context.Context, cap config.Capsule) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(cap.GetData())
	} else {
		check = cap.Get(c.Key).String()
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
		return false, fmt.Errorf("condition ip: type %s: %v", c.Type, ipInvalidType)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
