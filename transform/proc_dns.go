//go:build !wasm

package transform

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type procDNSConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// IgnoreErrors indicates if errors returned by the AWS Lambda function should be ignored.
	//
	// This is optional and defaults to false.
	IgnoreErrors bool `json:"ignore_errors"`
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

type procDNS struct {
	conf     procDNSConfig
	isObject bool

	resolver net.Resolver
	timeout  time.Duration
}

func newProcDNS(ctx context.Context, cfg config.Config) (*procDNS, error) {
	conf := procDNSConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") ||
		(conf.Key == "" && conf.SetKey != "") {
		return nil, fmt.Errorf("transform: proc_dns: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: proc_dns: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"forward_lookup",
			"reverse_lookup",
			"query_txt",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: proc_dns: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	if conf.Timeout == 0 {
		conf.Timeout = 1000
	}

	proc := procDNS{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
		resolver: net.Resolver{},
		timeout:  time.Duration(conf.Timeout) * time.Millisecond,
	}

	return &proc, nil
}

func (*procDNS) Close(context.Context) error {
	return nil
}

//nolint: gocognit // Ignore cognitive complexity.
func (t *procDNS) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	resolverCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel() // important to avoid a resource leak

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.isObject {
		case true:
			result := message.Get(t.conf.Key).String()

			switch t.conf.Type {
			case "forward_lookup":
				addrs, err := t.resolver.LookupHost(resolverCtx, result)
				if err != nil && !t.conf.IgnoreErrors {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				if err := message.Set(t.conf.SetKey, addrs); err != nil {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				output = append(output, message)
			case "reverse_lookup":
				names, err := t.resolver.LookupAddr(resolverCtx, result)
				if err != nil && !t.conf.IgnoreErrors {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				if err := message.Set(t.conf.SetKey, names); err != nil {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				output = append(output, message)
			case "query_txt":
				records, err := t.resolver.LookupTXT(resolverCtx, result)
				if err != nil && !t.conf.IgnoreErrors {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				if err := message.Set(t.conf.SetKey, records); err != nil {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				output = append(output, message)
			}

		case false:
			result := string(message.Data())

			switch t.conf.Type {
			case "forward_lookup":
				addrs, err := t.resolver.LookupHost(resolverCtx, result)
				if err != nil && !t.conf.IgnoreErrors {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				// Return the first address.
				msg, err := mess.New(
					mess.SetData([]byte(addrs[0])),
				)
				if err != nil {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				output = append(output, msg)
			case "reverse_lookup":
				names, err := t.resolver.LookupAddr(resolverCtx, result)
				if err != nil && !t.conf.IgnoreErrors {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				// Return the first name.
				msg, err := mess.New(
					mess.SetData([]byte(names[0])),
				)
				if err != nil {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				output = append(output, msg)
			case "query_txt":
				records, err := t.resolver.LookupTXT(resolverCtx, result)
				if err != nil && !t.conf.IgnoreErrors {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				// Return the first record.
				msg, err := mess.New(
					mess.SetData([]byte(records[0])),
				)
				if err != nil {
					return nil, fmt.Errorf("transform: proc_dns: %v", err)
				}

				output = append(output, msg)
			}
		}
	}

	return output, nil
}
