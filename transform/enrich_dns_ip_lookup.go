//go:build !wasm

package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newEnrichDNSIPLookup(_ context.Context, cfg config.Config) (*enrichDNSIPLookup, error) {
	conf := enrichDNSConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform enrich_dns_ip_lookup: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "enrich_dns_ip_lookup"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	dur, err := time.ParseDuration(conf.Request.Timeout)
	if err != nil {
		return nil, fmt.Errorf("transform %s: duration: %v", conf.ID, err)
	}

	tf := enrichDNSIPLookup{
		conf:     conf,
		isObj:    conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
		resolver: net.Resolver{},
		timeout:  dur,
	}

	return &tf, nil
}

type enrichDNSIPLookup struct {
	conf  enrichDNSConfig
	isObj bool

	resolver net.Resolver
	timeout  time.Duration
}

// Transform performs a DNS lookup on a message.
func (tf *enrichDNSIPLookup) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	resolverCtx, cancel := context.WithTimeout(ctx, tf.timeout)
	defer cancel() // important to avoid a resource leak

	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		addrs, err := tf.resolver.LookupAddr(resolverCtx, str)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		// Return the first address.
		data := []byte(addrs[0])
		msg.SetData(data)
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	addrs, err := tf.resolver.LookupAddr(resolverCtx, value.String())
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, addrs); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichDNSIPLookup) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
