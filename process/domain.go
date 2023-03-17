package process

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/condition"
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
type procDomain struct {
	process
	Options procDomainOptions `json:"options"`
}

type procDomainOptions struct {
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
func (p procDomain) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procDomain) Close(context.Context) error {
	return nil
}

// Create a new domain processor.
func newProcDomain(cfg config.Config) (p procDomain, err error) {
	err = config.Decode(cfg.Settings, &p)
	if err != nil {
		return procDomain{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
	if err != nil {
		return procDomain{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"tld",
			"domain",
			"subdomain",
		},
		p.Options.Type) {
		return procDomain{}, fmt.Errorf("process: domain: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// validate data processing pattern
	if (p.Key != "" && p.SetKey == "") ||
		(p.Key == "" && p.SetKey != "") {
		return procDomain{}, fmt.Errorf("process: domain: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	return p, nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procDomain) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procDomain) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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

func (p procDomain) domain(s string) (string, error) {
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
