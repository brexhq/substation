package process

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errdomainNoSubdomain is returned when a domain without a subdomain is processed.
const errdomainNoSubdomain = errors.Error("no subdomain")

/*
domain processes data by parsing fully qualified domain names into labels. The processor supports these patterns:

	JSON:
		{"domain":"example.com"} >>> {"domain":"example.com","tld":"com"}
	data:
		example.com >>> com

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "domain",
		"settings": {
			"options": {
				"function": "tld"
			},
			"input_key": "domain",
			"output_key": "tld"
		}
	}
*/
type domain struct {
	process
	Options domainOptions `json:"options"`
}

/*
domainOptions contains custom options for the domain processor:

	Type:
		domain processing function applied to the data
		must be one of:
			tld
			domain
			subdomain
*/
type domainOptions struct {
	Type string `json:"type"`
}

// Close closes resources opened by the domain processor.
func (p domain) Close(context.Context) error {
	return nil
}

func (p domain) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process domain: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the domain processor.
func (p domain) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return capsule, fmt.Errorf("process domain: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()
		value, _ := p.domain(result)

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process domain: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		value, _ := p.domain(string(capsule.Data()))
		capsule.SetData([]byte(value))

		return capsule, nil
	}

	return capsule, fmt.Errorf("process domain: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}

func (p domain) domain(s string) (string, error) {
	switch p.Options.Type {
	case "tld":
		tld, _ := publicsuffix.PublicSuffix(s)
		return tld, nil
	case "domain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("process domain %s: %v", s, err)
		}
		return domain, nil
	case "subdomain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("process domain: %s: %v", s, err)
		}

		// subdomain is the input string minus the domain and a leading dot:
		// input == "foo.bar.com"
		// domain == "bar.com"
		// subdomain == "foo" ("foo.bar.com" minus ".bar.com")
		subdomain := strings.Replace(s, "."+domain, "", 1)
		if subdomain == domain {
			return "", fmt.Errorf("process domain %s: %v", s, errdomainNoSubdomain)
		}
		return subdomain, nil
	default:
		return "", nil
	}
}
