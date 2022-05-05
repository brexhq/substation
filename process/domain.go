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

// DomainInvalidSettings is returned when the Domain processor is configured with invalid Input and Output settings.
const DomainInvalidSettings = errors.Error("DomainInvalidSettings")

// DomainNoSubdomain is used when a domain without a subdomain is processed
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
	Function string `mapstructure:"function"`
}

/*
Domain processes data by parsing fully qualified domain names into labels. The processor supports these patterns:
	json:
		{"domain":"example.com"} >>> {"domain":"example.com","tld":"com"}
	json array:
		{"domain":["example.com","example.top"]} >>> {"domain":["example.com","example.top"],"tld":["com","top"]}
	data:
		example.com >>> com

The processor uses this Jsonnet configuration:
	{
		type: 'domain',
		settings: {
			input: {
				key: 'domain',
			},
			output: {
				key: 'tld',
			},
			options: {
				function: 'tld',
			}
		},
	}
*/
type Domain struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   DomainOptions            `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Domain processor. Conditions are optionally applied on the channel data to enable processing.
func (p Domain) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	var array [][]byte
	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil

}

// Byte processes a byte slice with the Domain processor.
func (p Domain) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// json processing
	if p.Input.Key != "" && p.Output.Key != "" {
		value := json.Get(data, p.Input.Key)
		if !value.IsArray() {
			label, _ := p.domain(value.String())
			return json.Set(data, p.Output.Key, label)
		}

		// json array processing
		var array []string
		for _, v := range value.Array() {
			label, _ := p.domain(v.String())
			array = append(array, label)
		}

		return json.Set(data, p.Output.Key, array)
	}

	// data processing
	if p.Input.Key == "" && p.Output.Key == "" {
		label, _ := p.domain(string(data))
		return []byte(label), nil
	}

	return nil, DomainInvalidSettings
}

func (p Domain) domain(s string) (string, error) {
	switch f := p.Options.Function; f {
	case "tld":
		tld, _ := publicsuffix.PublicSuffix(s)
		return tld, nil
	case "domain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("err Domain processor failed to parse domain from %s: %v", s, DomainNoSubdomain)
		}
		return domain, nil
	case "subdomain":
		domain, err := publicsuffix.EffectiveTLDPlusOne(s)
		if err != nil {
			return "", fmt.Errorf("err Domain processor failed to parse domain from %s: %v", s, DomainNoSubdomain)
		}

		// subdomain is the input string minus the domain and a leading dot
		// 	input == "foo.bar.com"
		// 	domain == "bar.com"
		// 	subdomain == "foo" ("foo.bar.com" minus ".bar.com")
		subdomain := strings.Replace(s, "."+domain, "", 1)
		if subdomain == domain {
			return "", fmt.Errorf("err Domain processor failed to parse subdomain from %s: %v", s, DomainNoSubdomain)
		}
		return subdomain, nil
	default:
		return "", nil
	}
}
