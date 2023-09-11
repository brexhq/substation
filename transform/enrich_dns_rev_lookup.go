//go:build !wasm

package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newEnrichDNSRevLookup(_ context.Context, cfg config.Config) (*enrichDNSRevLookup, error) {
	conf := enrichDNSConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_enrich_dns_rev_lookup: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_enrich_dns_rev_lookup: %v", err)
	}

	dur, err := time.ParseDuration(conf.Request.Timeout)
	if err != nil {
		return nil, fmt.Errorf("transform: new_enrich_dns_rev_lookup: duration: %v", err)
	}

	tf := enrichDNSRevLookup{
		conf:     conf,
		isObj:    conf.Object.Key != "" && conf.Object.SetKey != "",
		resolver: net.Resolver{},
		timeout:  dur,
	}

	return &tf, nil
}

type enrichDNSRevLookup struct {
	conf  enrichDNSConfig
	isObj bool

	resolver net.Resolver
	timeout  time.Duration
}

// Transform performs a DNS lookup on a message.
func (tf *enrichDNSRevLookup) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	resolverCtx, cancel := context.WithTimeout(ctx, tf.timeout)
	defer cancel() // important to avoid a resource leak

	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		names, err := tf.resolver.LookupAddr(resolverCtx, str)
		if err != nil {
			return nil, fmt.Errorf("transform: enrich_dns_rev_lookup: %v", err)
		}

		// Return the first name.
		data := []byte(names[0])
		msg.SetData(data)
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	names, err := tf.resolver.LookupAddr(resolverCtx, value.String())
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_dns_rev_lookup: %v", err)
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, names); err != nil {
		return nil, fmt.Errorf("transform: enrich_dns_rev_lookup: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichDNSRevLookup) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
