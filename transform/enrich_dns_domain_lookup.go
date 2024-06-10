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

func newEnrichDNSDomainLookup(_ context.Context, cfg config.Config) (*enrichDNSDomainLookup, error) {
	conf := enrichDNSConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform enrich_dns_domain_lookup: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "enrich_dns_domain_lookup"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	dur, err := time.ParseDuration(conf.Request.Timeout)
	if err != nil {
		return nil, fmt.Errorf("transform %s: duration: %v", conf.ID, err)
	}

	tf := enrichDNSDomainLookup{
		conf:     conf,
		isObj:    conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
		resolver: net.Resolver{},
		timeout:  dur,
	}

	return &tf, nil
}

type enrichDNSDomainLookup struct {
	conf  enrichDNSConfig
	isObj bool

	resolver net.Resolver
	timeout  time.Duration
}

// Transform performs a DNS lookup on a message.
func (tf *enrichDNSDomainLookup) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	resolverCtx, cancel := context.WithTimeout(ctx, tf.timeout)
	defer cancel() // important to avoid a resource leak

	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		names, err := tf.resolver.LookupHost(resolverCtx, str)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		// Return the first name.
		data := []byte(names[0])
		msg.SetData(data)
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	names, err := tf.resolver.LookupHost(resolverCtx, value.String())
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, names); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichDNSDomainLookup) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
