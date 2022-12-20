package condition

import (
	"context"
	"fmt"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errIPInvalidType is returned when the ip inspector is configured with an invalid type.
const errIPInvalidType = errors.Error("invalid type")

// ip evaluates IP addresses by their type and usage using the standard library's net package (more information is available here: https://pkg.go.dev/net#ip).
//
// This inspector supports the data and object handling patterns.
type _ip struct {
	condition
	Options _ipOptions `json:"options"`
}

type _ipOptions struct {
	// Type is the IP address type used for comparison during inspection.
	//
	// Must be one of:
	//
	// - valid: valid address of any type
	//
	// - loopback: valid loopback address
	//
	// - multicast: valid multicast address
	//
	// - multicast_link_local: valid link local multicast address
	//
	// - private: valid private address
	//
	// - unicast_global: valid global unicast address
	//
	// - unicast_link_local: valid link local unicast address
	//
	// - unspecified: valid "unspecified" address (e.g., 0.0.0.0, ::)
	Type string `json:"type"`
}

func (c _ip) String() string {
	return inspectorToString(c)
}

// Inspect evaluates encapsulated data with the ip inspector.
func (c _ip) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(capsule.Data())
	} else {
		check = capsule.Get(c.Key).String()
	}

	ip := net.ParseIP(check)
	var matched bool
	switch c.Options.Type {
	case "valid":
		matched = ip != nil
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
		return false, fmt.Errorf("condition ip: type %s: %v", c.Options.Type, errIPInvalidType)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
