package condition

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// ip evaluates IP addresses by their type and usage using the standard library's net package (more information is available here: https://pkg.go.dev/net#ip).
//
// This inspector supports the data and object handling patterns.
type inspIP struct {
	condition
	Options inspIPOptions `json:"options"`
}

type inspIPOptions struct {
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

// Creates a new IP inspector.
func newInspIP(_ context.Context, cfg config.Config) (c inspIP, err error) {
	if err = config.Decode(cfg.Settings, &c); err != nil {
		return inspIP{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"valid",
			"loopback",
			"multicast",
			"multicast_link_local",
			"private",
			"unicast_global",
			"unicast_link_local",
			"unspecified",
		},
		c.Options.Type) {
		return inspIP{}, fmt.Errorf("condition: ip: type %q: %v", c.Options.Type, errors.ErrInvalidOption)
	}

	return c, nil
}

func (c inspIP) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the ip inspector.
func (c inspIP) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
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
		return false, fmt.Errorf("condition: ip: type %s: %v", c.Options.Type, errors.ErrInvalidOption)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
