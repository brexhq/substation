package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newNetworkDomainRegisteredDomain(_ context.Context, cfg config.Config) (*networkDomainRegisteredDomain, error) {
	conf := networkDomainConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform network_domain_registered_domain: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "network_domain_registered_domain"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := networkDomainRegisteredDomain{
		conf:  conf,
		isObj: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type networkDomainRegisteredDomain struct {
	conf  networkDomainConfig
	isObj bool
}

func (tf *networkDomainRegisteredDomain) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		domain, err := publicsuffix.EffectiveTLDPlusOne(str)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		msg.SetData([]byte(domain))
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	domain, err := publicsuffix.EffectiveTLDPlusOne(value.String())
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, domain); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *networkDomainRegisteredDomain) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
