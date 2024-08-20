package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

// errFmtSubdomainNoSubdomain is returned when a domain without a subdomain is
// processed.
var errFmtSubdomainNoSubdomain = fmt.Errorf("no subdomain")

func newNetworkDomainSubdomain(_ context.Context, cfg config.Config) (*networkDomainSubdomain, error) {
	conf := networkDomainConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform network_domain_subdomain: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "network_domain_subdomain"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := networkDomainSubdomain{
		conf:  conf,
		isObj: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type networkDomainSubdomain struct {
	conf  networkDomainConfig
	isObj bool
}

func (tf *networkDomainSubdomain) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		domain, err := fmtParseSubdomain(str)
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

	domain, err := fmtParseSubdomain(value.String())
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, domain); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *networkDomainSubdomain) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func fmtParseSubdomain(s string) (string, error) {
	domain, err := publicsuffix.EffectiveTLDPlusOne(s)
	if err != nil {
		return "", err
	}

	// Subdomain is the input string minus the domain and a leading dot:
	// input == "foo.bar.com"
	// domain == "bar.com"
	// subdomain == "foo" ("foo.bar.com" minus ".bar.com")
	subdomain := strings.Replace(s, "."+domain, "", 1)
	if subdomain == domain {
		return "", errFmtSubdomainNoSubdomain
	}

	return subdomain, nil
}
