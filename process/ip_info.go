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
IPInfo processes data by querying it against IP address enrichment databases, including geographic location (geo) and autonomous system (asn) databases. The processor is designed to support enrichment databases from many service providers and each database is contextually loaded based on environment variables.

The processor abstracts information returned by enrichment databases into these categories:

- asn (autonomous system information)

- geo (location information)

The database is queried based on the naming convention [provider]_[asn|geo]. For example, maxmind_geo returns geolocation data from any available MaxMind database. The processor supports these database sources:

- IP2Location

- MaxMind ASN (GeoLite2)

- MaxMind City (GeoIP2 / GeoLite2)

The processor supports these patterns:

	JSON:
	  	{"ip":"8.8.8.8"} >>> {"ip":"8.8.8.8","as":{"number":15169,"organization":"GOOGLE"}}
	  	{"ip":"8.8.8.8"} >>> {"ip":"8.8.8.8","geo":{"continent":"North America","country":"United States","latitude":37.751,"longitude":-97.822,"accuracy_radius":1000,"timezone":"America/Chicago"}}

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "ip_info",
		"settings": {
			"options": {
				"function": "maxmind_geo"
			},
			"input_key": "ip",
			"output_key": "geo"
		}
	}
*/
type IPInfo struct {
	Options   IPInfoOptions    `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

// IPInfoOptions contains custom options for the DNS processor.
type IPInfoOptions struct {
	Function string `json:"function"`
}

// ApplyBatch processes a slice of encapsulated data with the DNS processor. Conditions are optionally applied to the data to enable processing.
func (p IPInfo) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process case: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process case: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the IPInfo processor.
func (p IPInfo) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return capsule, fmt.Errorf("process geoip: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
	}

	result := capsule.Get(p.InputKey).String()

	switch p.Options.Function {
	case "ip2location_geo":
		if !ipinfoIP2location.IsEnabled() {
			if err := ipinfoIP2location.Load(ctx); err != nil {
				return capsule, fmt.Errorf("process geoip: %v", err)
			}
		}

		resp, err := ipinfoIP2location.Location(result)
		if err != nil {
			return capsule, fmt.Errorf("process geoip: %v", err)
		}

		if err := capsule.Set(p.OutputKey, resp); err != nil {
			return capsule, fmt.Errorf("process whois: %v", err)
		}

	case "maxmind_asn":
		if !ipinfoMaxmind.IsASEnabled() {
			if err := ipinfoMaxmind.LoadAS(ctx); err != nil {
				return capsule, fmt.Errorf("process geoip: %v", err)
			}
		}

		resp, err := ipinfoMaxmind.AS(result)
		if err != nil {
			return capsule, fmt.Errorf("process geoip: %v", err)
		}

		if err := capsule.Set(p.OutputKey, resp); err != nil {
			return capsule, fmt.Errorf("process whois: %v", err)
		}

	case "maxmind_geo":
		if !ipinfoMaxmind.IsGeoEnabled() {
			if err := ipinfoMaxmind.LoadGeo(ctx); err != nil {
				return capsule, fmt.Errorf("process geoip: %v", err)
			}
		}

		resp, err := ipinfoMaxmind.Location(result)
		if err != nil {
			return capsule, fmt.Errorf("process geoip: %v", err)
		}

		if err := capsule.Set(p.OutputKey, resp); err != nil {
			return capsule, fmt.Errorf("process whois: %v", err)
		}
	}

	return capsule, nil
}
