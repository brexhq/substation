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

// DomainNoSubdomain is used when a domain without a subdomain is processed
const DomainNoSubdomain = errors.Error("DomainNoSubdomain")

/*
DomainOptions contain custom options settings for this processor.

Function: the domain processing function to apply to the data; one of: tld, domain, or subdomain
*/
type DomainOptions struct {
	Function string `mapstructure:"function"`
}

// Domain implements the Byter and Channeler interfaces and parses fully qualified domain names into separate labels. More information is available in the README.
type Domain struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   DomainOptions            `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Domain) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

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

// Byte processes a byte slice with this processor
func (p Domain) Byte(ctx context.Context, data []byte) ([]byte, error) {
	value := json.Get(data, p.Input.Key)

	if !value.IsArray() {
		s := value.String()
		output, _ := p.domain(s)

		if output == "" {
			return data, nil
		}

		return json.Set(data, p.Output.Key, output)
	}

	var array []string
	for _, v := range value.Array() {
		s := v.String()
		o, _ := p.domain(s)
		array = append(array, o)
	}

	return json.Set(data, p.Output.Key, array)
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
