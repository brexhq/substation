package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newNetworkDomainTopLevelDomain(_ context.Context, cfg config.Config) (*networkDomainTopLevelDomain, error) {
	conf := networkDomainConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform network_domain_top_level_domain: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "network_domain_top_level_domain"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := networkDomainTopLevelDomain{
		conf:  conf,
		isObj: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type networkDomainTopLevelDomain struct {
	conf  networkDomainConfig
	isObj bool
}

func (tf *networkDomainTopLevelDomain) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		domain, _ := publicsuffix.PublicSuffix(str)

		msg.SetData([]byte(domain))
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	domain, _ := publicsuffix.PublicSuffix(value.String())

	if err := msg.SetValue(tf.conf.Object.TargetKey, domain); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *networkDomainTopLevelDomain) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
