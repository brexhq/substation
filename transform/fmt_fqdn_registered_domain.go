package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newFmtFQDNRegisteredDomain(_ context.Context, cfg config.Config) (*fmtFQDNRegisteredDomain, error) {
	conf := fmtFQDNConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_fqdn_registered_domain: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_fqdn_registered_domain: %v", err)
	}

	tf := fmtFQDNRegisteredDomain{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type fmtFQDNRegisteredDomain struct {
	conf  fmtFQDNConfig
	isObj bool
}

func (tf *fmtFQDNRegisteredDomain) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		domain, err := publicsuffix.EffectiveTLDPlusOne(str)
		if err != nil {
			return nil, fmt.Errorf("transform: fmt_fqdn_registered_domain: %v", err)
		}

		msg.SetData([]byte(domain))
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	domain, err := publicsuffix.EffectiveTLDPlusOne(value.String())
	if err != nil {
		return nil, fmt.Errorf("transform: fmt_fqdn_registered_domain: %v", err)
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, domain); err != nil {
		return nil, fmt.Errorf("transform: fmt_fqdn_registered_domain: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *fmtFQDNRegisteredDomain) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*fmtFQDNRegisteredDomain) Close(context.Context) error {
	return nil
}
