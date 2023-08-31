package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// errModDomainNoSubdomain is returned when a domain without a subdomain is
// processed.
var errModDomainNoSubdomain = fmt.Errorf("no subdomain")

type modDomainConfig struct {
	Object configObject `json:"object"`

	// ErrorOnFailure determines whether an error is returned during processing.
	//
	// This is optional and defaults to false.
	ErrorOnFailure bool `json:"error_on_failure"`
	// Type is the domain function applied to the data.
	//
	// Must be one of:
	//	- top_level_domain
	//	- registered_domain
	//	- subdomain
	Type string `json:"type"`
}

type modDomain struct {
	conf  modDomainConfig
	isObj bool
}

func newModDomain(_ context.Context, cfg config.Config) (*modDomain, error) {
	conf := modDomainConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_domain: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_domain: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_domain: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: new_mod_domain: type: %v", errors.ErrMissingRequiredOption)
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
		return nil, fmt.Errorf("transform: new_mod_domain: options %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	tf := modDomain{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *modDomain) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modDomain) Close(context.Context) error {
	return nil
}

func (tf *modDomain) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		result := string(msg.Data())
		value, err := tf.process(result)
		// If ErrorOnFailure is configured, then errors are returned,
		// but otherwise the message is returned as-is.
		if err != nil && tf.conf.ErrorOnFailure {
			return nil, fmt.Errorf("transform: mod_domain: %v", err)
		} else if err != nil {
			//nolint: nilerr // err is configurable.
			return []*message.Message{msg}, nil
		}

		data := []byte(value)
		finMsg := message.New().SetData(data).SetMetadata(msg.Metadata())
		return []*message.Message{finMsg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key).String()
	value, err := tf.process(result)

	// If ErrorOnFailure is configured, then errors are returned,
	// but otherwise the message is returned as-is.
	if err != nil && tf.conf.ErrorOnFailure {
		return nil, fmt.Errorf("transform: mod_domain: %v", err)
	} else if err != nil {
		//nolint: nilerr // err is configurable.
		return []*message.Message{msg}, nil
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_domain: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *modDomain) process(s string) (string, error) {
	switch tf.conf.Type {
	case "tld", "top_level_domain":
		tld, _ := publicsuffix.PublicSuffix(s)
		return tld, nil
	case "domain", "registered_domain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", err
		}
		return domain, nil
	case "subdomain":
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
			return "", errModDomainNoSubdomain
		}

		return subdomain, nil
	}

	return "", errors.ErrInvalidOption
}
