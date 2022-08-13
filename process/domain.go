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

// DomainNoSubdomain is returned when a domain without a subdomain is processed.
const DomainNoSubdomain = errors.Error("DomainNoSubdomain")

/*
DomainOptions contains custom options for the Domain processor:
	Function:
		the domain processing function to apply to the data
		must be one of:
			tld
			domain
			subdomain
*/
type DomainOptions struct {
	Function string `json:"function"`
}

/*
Domain processes encapsulated data by parsing fully qualified domain names into labels. The processor supports these patterns:
	JSON:
		{"domain":"example.com"} >>> {"domain":"example.com","tld":"com"}
	data:
		example.com >>> com

The processor uses this Jsonnet configuration:
	{
		type: 'domain',
		settings: {
			input_key: 'domain',
			input_key: 'tld',
			options: {
				_function: 'tld',
			}
		},
	}
*/
type Domain struct {
	Options   DomainOptions            `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Domain processor. Conditions are optionally applied to the data to enable processing.
func (p Domain) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Domain processor.
func (p Domain) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Function == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		res := cap.Get(p.InputKey).String()
		label, _ := p.domain(res)

		cap.Set(p.OutputKey, label)
		return cap, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		label, _ := p.domain(string(cap.GetData()))

		cap.SetData([]byte(label))
		return cap, nil
	}

	return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
}

func (p Domain) domain(s string) (string, error) {
	switch p.Options.Function {
	case "tld":
		tld, _ := publicsuffix.PublicSuffix(s)
		return tld, nil
	case "domain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("domain %s: %v", s, DomainNoSubdomain)
		}
		return domain, nil
	case "subdomain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("domain %s: %v", s, DomainNoSubdomain)
		}

		// subdomain is the input string minus the domain and a leading dot:
		// input == "foo.bar.com"
		// domain == "bar.com"
		// subdomain == "foo" ("foo.bar.com" minus ".bar.com")
		subdomain := strings.Replace(s, "."+domain, "", 1)
		if subdomain == domain {
			return "", fmt.Errorf("domain %s: %v", s, DomainNoSubdomain)
		}
		return subdomain, nil
	default:
		return "", nil
	}
}
