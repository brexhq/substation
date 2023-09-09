package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

// errFmtSubdomainNoSubdomain is returned when a domain without a subdomain is
// processed.
var errFmtSubdomainNoSubdomain = fmt.Errorf("no subdomain")

func newFmtFQDNSubdomain(_ context.Context, cfg config.Config) (*fmtFQDNSubdomain, error) {
	conf := fmtFQDNConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_fqdn_subdomain: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_fqdn_subdomain: %v", err)
	}

	tf := fmtFQDNSubdomain{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type fmtFQDNSubdomain struct {
	conf  fmtFQDNConfig
	isObj bool
}

func (tf *fmtFQDNSubdomain) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		str := string(msg.Data())
		domain, err := fmtParseSubdomain(str)
		if err != nil {
			return nil, fmt.Errorf("transform: fmt_fqdn_subdomain: %v", err)
		}

		msg.SetData([]byte(domain))
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	domain, err := fmtParseSubdomain(value.String())
	if err != nil {
		return nil, fmt.Errorf("transform: fmt_fqdn_subdomain: %v", err)
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, domain); err != nil {
		return nil, fmt.Errorf("transform: fmt_fqdn_subdomain: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *fmtFQDNSubdomain) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*fmtFQDNSubdomain) Close(context.Context) error {
	return nil
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
