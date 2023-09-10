package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkFQDNTLD(_ context.Context, cfg config.Config) (*networkFQDNTLD, error) {
	conf := networkFQDNConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_fqdn_tld: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_fqdn_tld: %v", err)
	}

	tf := networkFQDNTLD{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type networkFQDNTLD struct {
	conf  networkFQDNConfig
	isObj bool
}

func (tf *networkFQDNTLD) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		domain, _ := publicsuffix.PublicSuffix(str)

		msg.SetData([]byte(domain))
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	domain, _ := publicsuffix.PublicSuffix(value.String())

	if err := msg.SetValue(tf.conf.Object.SetKey, domain); err != nil {
		return nil, fmt.Errorf("transform: fmt_fqdn_tld: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *networkFQDNTLD) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*networkFQDNTLD) Close(context.Context) error {
	return nil
}
