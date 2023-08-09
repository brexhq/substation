package condition

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"net"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type inspIPConf struct {
	// Key is the message key used during inspection.
	Key string `json:"key"`
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
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

type inspIP struct {
	conf inspIPConf
}

func newInspIP(_ context.Context, cfg config.Config) (*inspIP, error) {
	conf := inspIPConf{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Type == "" {
		return nil, fmt.Errorf("condition: insp_ip: type: %v", errors.ErrMissingRequiredOption)
	}

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
		conf.Type) {
		return nil, fmt.Errorf("condition: insp_ip: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	sink := inspIP{
		conf: conf,
	}

	return &sink, nil
}

func (c *inspIP) String() string {
	b, _ := gojson.Marshal(c.conf)
	return string(b)
}

func (c *inspIP) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	var check string
	if c.conf.Key == "" {
		check = string(message.Data())
	} else {
		check = message.Get(c.conf.Key).String()
	}

	ip := net.ParseIP(check)
	var matched bool
	switch c.conf.Type {
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
	}

	if c.conf.Negate {
		return !matched, nil
	}

	return matched, nil
}
