package process

import (
	"context"
	"fmt"
	"net"
	gotime "time"

	"github.com/brexhq/substation/config"
)

var dnsResolver net.Resolver

type dns struct {
	process
	Options dnsOptions `json:"options"`
}

type dnsOptions struct {
	Type    string `json:"type"`
	Timeout int    `json:"timeout"`
}

// Close closes resources opened by the DNS processor.
func (p dns) Close(context.Context) error {
	return nil
}

func (p dns) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process capture: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the DNS processor.
func (p dns) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return capsule, fmt.Errorf("process dns: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	var timeout gotime.Duration
	if p.Options.Timeout != 0 {
		timeout = gotime.Duration(p.Options.Timeout) * gotime.Millisecond
	} else {
		timeout = gotime.Duration(1000) * gotime.Millisecond
	}

	resolverCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // important to avoid a resource leak

	// JSON processing
	//nolint: nestif // ignore nesting
	if p.Key != "" && p.SetKey != "" {
		res := capsule.Get(p.Key).String()

		switch p.Options.Type {
		case "forward_lookup":
			addrs, err := dnsResolver.LookupHost(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			if err := capsule.Set(p.SetKey, addrs); err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			return capsule, nil
		case "reverse_lookup":
			names, err := dnsResolver.LookupAddr(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			if err := capsule.Set(p.SetKey, names); err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			return capsule, nil
		case "query_txt":
			records, err := dnsResolver.LookupTXT(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			if err := capsule.Set(p.SetKey, records); err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			return capsule, nil
		default:
			return capsule, nil
		}
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		res := string(capsule.Data())

		switch p.Options.Type {
		case "forward_lookup":
			addrs, err := dnsResolver.LookupHost(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			// can only return one value, which is the first address
			capsule.SetData([]byte(addrs[0]))

			return capsule, nil
		case "reverse_lookup":
			names, err := dnsResolver.LookupAddr(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			// can only return one value, which is the first name
			capsule.SetData([]byte(names[0]))
			return capsule, nil
		case "query_txt":
			records, err := dnsResolver.LookupTXT(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			// can only return one value, which is the first record
			capsule.SetData([]byte(records[0]))
			return capsule, nil
		default:
			return capsule, nil
		}
	}

	return capsule, fmt.Errorf("process dns: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
