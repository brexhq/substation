package process

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errDomainNoSubdomain is returned when a domain without a subdomain is processed.
const errDomainNoSubdomain = errors.Error("no subdomain")

/*
Domain processes data by parsing fully qualified domain names into labels. The processor supports these patterns:
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
type Domain struct {
	Options   DomainOptions    `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
DomainOptions contains custom options for the Domain processor:
	Function:
		the domain processing function to to the data
		must be one of:
			tld
			domain
			subdomain
*/
type DomainOptions struct {
	Function string `json:"function"`
}

// ApplyBatch processes a slice of encapsulated data with the Domain processor. Conditions are optionally applied to the data to enable processing.
func (p Domain) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process domain: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process domain: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Domain processor.
func (p Domain) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Function == "" {
		return cap, fmt.Errorf("process domain: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := cap.Get(p.InputKey).String()
		value, _ := p.domain(result)

		if err := cap.Set(p.OutputKey, value); err != nil {
			return cap, fmt.Errorf("process domain: %v", err)
		}

		return cap, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		value, _ := p.domain(string(cap.Data()))
		cap.SetData([]byte(value))

		return cap, nil
	}

	return cap, fmt.Errorf("process domain: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
}

func (p Domain) domain(s string) (string, error) {
	switch p.Options.Function {
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
			return "", fmt.Errorf("process domain %s: %v", s, errDomainNoSubdomain)
		}
		return subdomain, nil
	default:
		return "", nil
	}
}
