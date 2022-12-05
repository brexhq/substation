package process

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

var dnsResolver net.Resolver

/*
DNS processes data by querying domains or IP addresses in the Domain Name System (DNS). By default, this processor can take up to 1 second per DNS query and may have significant impact on end-to-end data processing latency. If Substation is running in AWS Lambda with Kinesis, then this latency can be mitigated by increasing the parallelization factor of the Lambda (https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html).

The processor supports these patterns:

	JSON:
	  	{"ip":"8.8.8.8"} >>> {"ip":"8.8.8.8","domains":["dns.google."]}
		{"domain":"dns.google"} >>> {"domain":"dns.google","ips":["8.8.4.4","8.8.8.8","2001:4860:4860::8844","2001:4860:4860::8888"]}
	data:
		8.8.8.8 >>> dns.google.
		dns.google >>> 8.8.4.4

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "dns",
		"settings": {
			"options": {
				"function": "reverse_lookup"
			},
			"input_key": "ip",
			"output_key": "domains"
		}
	}
*/
type DNS struct {
	Options   DNSOptions       `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
DNSOptions contains custom options for the DNS processor.

	Function:
		Type of query made to DNS.

		Must be one of:
			forward_lookup (retrieve IP addresses associated with a domain)
			reverse_lookup (retrieve domains associated with an IP address)
			query_txt (retrieve TXT records for a domain)

	Timeout (optional):
		Amount of time to wait (in milliseconds) for a response.

		Defaults to 1000 milliseconds (1 second).
*/
type DNSOptions struct {
	Function string `json:"function"`
	Timeout  int    `json:"timeout"`
}

// Close closes resources opened by the DNS processor.
func (p DNS) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the DNS processor. Conditions are optionally applied to the data to enable processing.
func (p DNS) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process dns: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process dns: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the DNS processor.
func (p DNS) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Function == "" {
		return capsule, fmt.Errorf("process dns: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	var timeout time.Duration
	if p.Options.Timeout != 0 {
		timeout = time.Duration(p.Options.Timeout) * time.Millisecond
	} else {
		timeout = time.Duration(1000) * time.Millisecond
	}

	resolverCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // important to avoid a resource leak

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		res := capsule.Get(p.InputKey).String()

		switch p.Options.Function {
		case "forward_lookup":
			addrs, err := dnsResolver.LookupHost(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			if err := capsule.Set(p.OutputKey, addrs); err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			return capsule, nil
		case "reverse_lookup":
			names, err := dnsResolver.LookupAddr(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			if err := capsule.Set(p.OutputKey, names); err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			return capsule, nil
		case "query_txt":
			records, err := dnsResolver.LookupTXT(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			if err := capsule.Set(p.OutputKey, records); err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			return capsule, nil
		default:
			return capsule, nil
		}
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		res := string(capsule.Data())

		switch p.Options.Function {
		case "forwardlookup":
			addrs, err := dnsResolver.LookupHost(resolverCtx, res)
			if err != nil {
				return capsule, fmt.Errorf("process dns: %v", err)
			}

			// can only return one value, which is the first address
			capsule.SetData([]byte(addrs[0]))

			return capsule, nil
		case "reverselookup":
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

	return capsule, fmt.Errorf("process dns: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
}
