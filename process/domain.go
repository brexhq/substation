package process

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errDomainNoSubdomain is returned when a domain without a subdomain is
// processed.
const errDomainNoSubdomain = errors.Error("no subdomain")

// domain processes data by parsing fully qualified domain names (FQDNs) into
// labels.
//
// This processor supports the data and object handling patterns.
type _domain struct {
	process
	Options _domainOptions `json:"options"`
}

type _domainOptions struct {
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

// String returns the processor settings as an object.
func (p _domain) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _domain) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _domain) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p _domain) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return capsule, fmt.Errorf("process: domain: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()
		value, _ := p.domain(result)

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process: domain: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		value, _ := p.domain(string(capsule.Data()))
		capsule.SetData([]byte(value))

		return capsule, nil
	}

	return capsule, fmt.Errorf("process: domain: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}

func (p _domain) domain(s string) (string, error) {
	switch p.Options.Type {
	case "tld":
		tld, _ := publicsuffix.PublicSuffix(s)
		return tld, nil
	case "domain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("process: domain %s: %v", s, err)
		}
		return domain, nil
	case "subdomain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("process: domain: %s: %v", s, err)
		}

		// subdomain is the input string minus the domain and a leading dot:
		// input == "foo.bar.com"
		// domain == "bar.com"
		// subdomain == "foo" ("foo.bar.com" minus ".bar.com")
		subdomain := strings.Replace(s, "."+domain, "", 1)
		if subdomain == domain {
			return "", fmt.Errorf("process: domain %s: %v", s, errDomainNoSubdomain)
		}
		return subdomain, nil
	default:
		return "", nil
	}
}
