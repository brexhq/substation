//go:build !wasm

package transform

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modDNSConfig struct {
	Object  configObject  `json:"object"`
	Request configRequest `json:"request"`

	// ErrorOnFailure determines whether an error is returned during processing.
	//
	// This is optional and defaults to false.
	ErrorOnFailure bool `json:"error_on_failure"`
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
}

type modDNS struct {
	conf  modDNSConfig
	isObj bool

	resolver net.Resolver
	timeout  time.Duration
}

func newModDNS(ctx context.Context, cfg config.Config) (*modDNS, error) {
	conf := modDNSConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_dns: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_dns: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_dns: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: new_mod_dns: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"forward_lookup",
			"reverse_lookup",
			"query_txt",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: new_mod_dns: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	if conf.Request.Timeout == "" {
		conf.Request.Timeout = "1s"
	}

	dur, err := time.ParseDuration(conf.Request.Timeout)
	if err != nil {
		return nil, fmt.Errorf("transform: new_mod_dns: duration: %v", err)
	}

	tf := modDNS{
		conf:     conf,
		isObj:    conf.Object.Key != "" && conf.Object.SetKey != "",
		resolver: net.Resolver{},
		timeout:  dur,
	}

	return &tf, nil
}

func (*modDNS) Close(context.Context) error {
	return nil
}

// Transform performs a DNS lookup on a message.
func (tf *modDNS) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	resolverCtx, cancel := context.WithTimeout(ctx, tf.timeout)
	defer cancel() // important to avoid a resource leak

	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		result := string(msg.Data())

		switch tf.conf.Type {
		case "forward_lookup":
			addrs, err := tf.resolver.LookupHost(resolverCtx, result)

			// If ErrorOnFailure is configured, then errors are returned,
			// but otherwise the message is returned as-is.
			if err != nil && tf.conf.ErrorOnFailure {
				return nil, fmt.Errorf("transform: mod_dns: %v", err)
			} else if err != nil {
				//nolint: nilerr // err is configurable.
				return []*message.Message{msg}, nil
			}

			// Return the first address.

			data := []byte(addrs[0])
			finMsg := message.New().SetData(data)
			return []*message.Message{finMsg}, nil
		case "reverse_lookup":
			names, err := tf.resolver.LookupAddr(resolverCtx, result)

			// If ErrorOnFailure is configured, then errors are returned,
			// but otherwise the message is returned as-is.
			if err != nil && tf.conf.ErrorOnFailure {
				return nil, fmt.Errorf("transform: mod_dns: %v", err)
			} else if err != nil {
				//nolint: nilerr // err is configurable.
				return []*message.Message{msg}, nil
			}

			// Return the first name.
			data := []byte(names[0])
			finMsg := message.New().SetData(data)
			return []*message.Message{finMsg}, nil
		case "query_txt":
			records, err := tf.resolver.LookupTXT(resolverCtx, result)

			// If ErrorOnFailure is configured, then errors are returned,
			// but otherwise the message is returned as-is.
			if err != nil && tf.conf.ErrorOnFailure {
				return nil, fmt.Errorf("transform: mod_dns: %v", err)
			} else if err != nil {
				//nolint: nilerr // err is configurable.
				return []*message.Message{msg}, nil
			}

			// Return the first record.
			data := []byte(records[0])
			finMsg := message.New().SetData(data)
			return []*message.Message{finMsg}, nil
		}
	}

	result := msg.GetObject(tf.conf.Object.Key).String()

	switch tf.conf.Type {
	case "forward_lookup":
		addrs, err := tf.resolver.LookupHost(resolverCtx, result)

		// If ErrorOnFailure is configured, then errors are returned,
		// but otherwise the message is returned as-is.
		if err != nil && tf.conf.ErrorOnFailure {
			return nil, fmt.Errorf("transform: mod_dns: %v", err)
		} else if err != nil {
			//nolint: nilerr // err is configurable.
			return []*message.Message{msg}, nil
		}

		if err := msg.SetObject(tf.conf.Object.SetKey, addrs); err != nil {
			return nil, fmt.Errorf("transform: mod_dns: %v", err)
		}
	case "reverse_lookup":
		names, err := tf.resolver.LookupAddr(resolverCtx, result)

		// If ErrorOnFailure is configured, then errors are returned,
		// but otherwise the message is returned as-is.
		if err != nil && tf.conf.ErrorOnFailure {
			return nil, fmt.Errorf("transform: mod_dns: %v", err)
		} else if err != nil {
			//nolint: nilerr // err is configurable.
			return []*message.Message{msg}, nil
		}

		if err := msg.SetObject(tf.conf.Object.SetKey, names); err != nil {
			return nil, fmt.Errorf("transform: mod_dns: %v", err)
		}
	case "query_txt":
		records, err := tf.resolver.LookupTXT(resolverCtx, result)

		// If ErrorOnFailure is configured, then errors are returned,
		// but otherwise the message is returned as-is.
		if err != nil && tf.conf.ErrorOnFailure {
			return nil, fmt.Errorf("transform: mod_dns: %v", err)
		} else if err != nil {
			//nolint: nilerr // err is configurable.
			return []*message.Message{msg}, nil
		}

		if err := msg.SetObject(tf.conf.Object.SetKey, records); err != nil {
			return nil, fmt.Errorf("transform: mod_dns: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}
