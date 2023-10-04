package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkDomainRegisteredDomain(_ context.Context, cfg config.Config) (*networkDomainRegisteredDomain, error) {
	conf := networkDomainConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: network_domain_registered_domain: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: network_domain_registered_domain: %v", err)
	}

	tf := networkDomainRegisteredDomain{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
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
			return nil, fmt.Errorf("transform: network_domain_registered_domain: %v", err)
		}

		msg.SetData([]byte(domain))
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	domain, err := publicsuffix.EffectiveTLDPlusOne(value.String())
	if err != nil {
		return nil, fmt.Errorf("transform: network_domain_registered_domain: %v", err)
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, domain); err != nil {
		return nil, fmt.Errorf("transform: network_domain_registered_domain: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *networkDomainRegisteredDomain) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
