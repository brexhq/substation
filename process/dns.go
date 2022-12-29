package process

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/brexhq/substation/config"
)

var dnsResolver net.Resolver

// dns processes data by querying domains or IP addresses in the Domain Name
// System (DNS). By default, this processor can take up to 1 second per DNS
// query and may have significant impact on end-to-end data processing latency.
// If Substation is running in AWS Lambda with Kinesis, then this latency can be
//
//	mitigated by increasing the parallelization factor of the Lambda
//
// (https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html).
type _dns struct {
	process
	Options _dnsOptions `json:"options"`
}

type _dnsOptions struct {
	// Type is the query type made to DNS.
	//
	// Must be one of:
	//
	// - forward_lookup: retrieve IP addresses associated with a domain
	//
	// - reverse_lookup: retrieve domains associated with an IP address
	//
	// - query_txt: retrieve TXT records for a domain
	Type string `json:"type"`
	// Timeout is the amount of time to wait (in milliseconds) for
	// a response.
	//
	// This is optional and defaults to 1000 milliseconds (1 second).
	Timeout int `json:"timeout"`
}

// Close closes resources opened by the processor.
func (p _dns) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _dns) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
//
//nolint:gocognit
func (p _dns) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return capsule, fmt.Errorf("process dns: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	var timeout time.Duration
	if p.Options.Timeout != 0 {
		timeout = time.Duration(p.Options.Timeout) * time.Millisecond
	} else {
		timeout = 1000 * time.Millisecond
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
			if err != nil && p.IgnoreErrors {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			if err := capsule.Set(p.SetKey, addrs); err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			return capsule, nil
		case "reverse_lookup":
			names, err := dnsResolver.LookupAddr(resolverCtx, res)
			if err != nil && p.IgnoreErrors {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			if err := capsule.Set(p.SetKey, names); err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			return capsule, nil
		case "query_txt":
			records, err := dnsResolver.LookupTXT(resolverCtx, res)
			if err != nil && p.IgnoreErrors {
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
			if err != nil && p.IgnoreErrors {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			// can only return one value, which is the first address
			capsule.SetData([]byte(addrs[0]))

			return capsule, nil
		case "reverse_lookup":
			names, err := dnsResolver.LookupAddr(resolverCtx, res)
			if err != nil && p.IgnoreErrors {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			// can only return one value, which is the first name
			capsule.SetData([]byte(names[0]))
			return capsule, nil
		case "query_txt":
			records, err := dnsResolver.LookupTXT(resolverCtx, res)
			if err != nil && p.IgnoreErrors {
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
