package process

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
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
Domain processes data by parsing fully qualified domain names into labels. The processor supports these patterns:
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

// Slice processes a slice of bytes with the Domain processor. Conditions are optionally applied on the bytes to enable processing.
func (p Domain) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %v: %v", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Domain processor.
func (p Domain) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Function == "" {
		return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey)
		label, _ := p.domain(value.String())
		return json.Set(data, p.OutputKey, label)
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		label, _ := p.domain(string(data))
		return []byte(label), nil
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, ProcessorInvalidSettings)
}

func (p Domain) domain(s string) (string, error) {
	switch f := p.Options.Function; f {
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
