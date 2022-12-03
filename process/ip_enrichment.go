package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/ip"
)

var (
	ipinfoMaxmind     ip.MaxMind
	ipinfoIP2location ip.IP2Location
)

/*
IPEnrichment processes data by querying IP addresses against enrichment databases, including geographic location (geo) and autonomous system (asn) databases. The processor supports multiple database providers by contextually retrieving and loading databases using environment variables and can be reused if multiple databases need to be queried.

IP address information is abstracted from each enrichment database into these categories:

- asn (autonomous system information)

- geo (location information)

Enrichment databases are selected based on the naming convention [provider]_[asn|geo]. For example, maxmind_geo returns geolocation data from any available MaxMind database. These database providers are supported:

- IP2Location

- MaxMind ASN (GeoLite2)

- MaxMind City (GeoIP2 or GeoLite2)

The processor supports these patterns:

	JSON:
	  	{"ip":"8.8.8.8"} >>> {"ip":"8.8.8.8","as":{"number":15169,"organization":"GOOGLE"}}
	  	{"ip":"8.8.8.8"} >>> {"ip":"8.8.8.8","geo":{"continent":"North America","country":"United States","latitude":37.751,"longitude":-97.822,"accuracy_radius":1000,"timezone":"America/Chicago"}}

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "ip_enrichment",
		"settings": {
			"options": {
				"function": "maxmind_geo"
			},
			"input_key": "ip",
			"output_key": "geo"
		}
	}
*/
type IPEnrichment struct {
	Options   IPEnrichmentOptions `json:"options"`
	Condition condition.Config    `json:"condition"`
	InputKey  string              `json:"input_key"`
	OutputKey string              `json:"output_key"`
}

/*
IPEnrichmentOptions contains custom options for the IPEnrichment processor.

	Function:
		Selects the enrichment database queried by the processor.

		The database is contextually retrieved using an environment variable and lazy loaded on first invocation. Each environment variable should contain the location of the database, which can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL.

		Must be one of:
			ip2location_geo (IP2LOCATION_DB)
			maxmind_asn (MAXMIND_ASN_DB)
			maxmind_geo (MAXMIND_LOCATION_DB)
*/
type IPEnrichmentOptions struct {
	Function string `json:"function"`
}

// ApplyBatch processes a slice of encapsulated data with the IPEnrichment processor. Conditions are optionally applied to the data to enable processing.
func (p IPEnrichment) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process ip_enrichment: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process ip_enrichment: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the IPEnrichment processor.
func (p IPEnrichment) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return capsule, fmt.Errorf("process ip_enrichment: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
	}

	result := capsule.Get(p.InputKey).String()

	switch p.Options.Function {
	case "ip2location_geo":
		if !ipinfoIP2location.IsEnabled() {
			if err := ipinfoIP2location.Setup(ctx); err != nil {
				return capsule, fmt.Errorf("process ip_enrichment: %v", err)
			}
		}

		resp, err := ipinfoIP2location.Location(result)
		if err != nil {
			return capsule, fmt.Errorf("process ip_enrichment: %v", err)
		}

		if err := capsule.Set(p.OutputKey, resp); err != nil {
			return capsule, fmt.Errorf("process ip_enrichment: %v", err)
		}

	case "maxmind_asn":
		if !ipinfoMaxmind.IsASNEnabled() {
			if err := ipinfoMaxmind.Setup(ctx); err != nil {
				return capsule, fmt.Errorf("process ip_enrichment: %v", err)
			}
		}

		resp, err := ipinfoMaxmind.ASN(result)
		if err != nil {
			return capsule, fmt.Errorf("process ip_enrichment: %v", err)
		}

		if err := capsule.Set(p.OutputKey, resp); err != nil {
			return capsule, fmt.Errorf("process ip_enrichment: %v", err)
		}

	case "maxmind_geo":
		if !ipinfoMaxmind.IsLocationEnabled() {
			if err := ipinfoMaxmind.Setup(ctx); err != nil {
				return capsule, fmt.Errorf("process ip_enrichment: %v", err)
			}
		}

		resp, err := ipinfoMaxmind.Location(result)
		if err != nil {
			return capsule, fmt.Errorf("process ip_enrichment: %v", err)
		}

		if err := capsule.Set(p.OutputKey, resp); err != nil {
			return capsule, fmt.Errorf("process ip_enrichment: %v", err)
		}
	}

	return capsule, nil
}
