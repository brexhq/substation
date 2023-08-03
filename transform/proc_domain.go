package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// errProcDomainNoSubdomain is returned when a domain without a subdomain is
// processed.
var errProcDomainNoSubdomain = fmt.Errorf("no subdomain")

type procDomainConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// ErrorOnFailure determines whether an error is returned during processing.
	//
	// This is optional and defaults to false.
	ErrorOnFailure bool `json:"error_on_failure"`
	// Type is the domain function applied to the data.
	//
	// Must be one of:
	//
	// - tld: top-level domain
	//
	// - domain
	//
	// - subdomain
	Type string `json:"type"`
}

type procDomain struct {
	conf     procDomainConfig
	isObject bool
}

func newProcDomain(_ context.Context, cfg config.Config) (*procDomain, error) {
	conf := procDomainConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") ||
		(conf.Key == "" && conf.SetKey != "") {
		return nil, fmt.Errorf("transform: proc_dns: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: proc_domain: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"tld", // provides backwards compatibility
			"top_level_domain",
			"domain", // provides backwards compatibility
			"registered_domain",
			"subdomain",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: proc_domain: options %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	proc := procDomain{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
	}

	return &proc, nil
}

func (t *procDomain) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procDomain) Close(context.Context) error {
	return nil
}

func (t *procDomain) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.isObject {
		case true:
			result := message.Get(t.conf.Key).String()
			value, err := t.process(result)
			if err != nil && t.conf.ErrorOnFailure {
				return nil, fmt.Errorf("transform: proc_domain: %v", err)
			}

			if err := message.Set(t.conf.SetKey, value); err != nil {
				return nil, fmt.Errorf("transform: proc_domain: %v", err)
			}

			output = append(output, message)

		case false:
			value, _ := t.process(string(message.Data()))
			msg, err := mess.New(
				mess.SetData([]byte(value)),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_domain: %v", err)
			}

			output = append(output, msg)
		}
	}

	return output, nil
}

func (t *procDomain) process(s string) (string, error) {
	switch t.conf.Type {
	case "tld":
		fallthrough
	case "top_level_domain":
		tld, _ := publicsuffix.PublicSuffix(s)
		return tld, nil
	case "domain":
		fallthrough
	case "registered_domain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("process: domain %s: %v", s, err)
		}
		return domain, nil
	case "subdomain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("transform: proc_domain: %s: %v", s, err)
		}

		// Subdomain is the input string minus the domain and a leading dot:
		// input == "foo.bar.com"
		// domain == "bar.com"
		// subdomain == "foo" ("foo.bar.com" minus ".bar.com")
		subdomain := strings.Replace(s, "."+domain, "", 1)
		if subdomain == domain {
			return "", fmt.Errorf("process: domain %s: %v", s, errProcDomainNoSubdomain)
		}
		return subdomain, nil
	}

	return "", fmt.Errorf("transform: proc_domain: %v", errors.ErrInvalidOption)
}
