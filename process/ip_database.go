package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	ipdb "github.com/brexhq/substation/internal/ip/database"
)

/*
IPDatabase processes data by querying IP addresses in enrichment databases, including geographic location (geo) and autonomous system (asn) databases. The processor supports multiple database providers and can be reused if multiple databases need to be queried.

IP address information is abstracted from each enrichment database into a single record that contains these categories:

- asn (autonomous system information)

- geo (location information)

Enrichment databases are selected based on the naming convention [provider]_[database_name]. For example, maxmind_city returns information from any MaxMind City database. These database providers are supported:

- IP2Location

- MaxMind ASN (GeoLite2)

- MaxMind City (GeoIP2 or GeoLite2)

The processor supports these patterns:

	JSON:
	  	{"ip":"8.8.8.8"} >>> {"ip":"8.8.8.8","as":{"number":15169,"organization":"GOOGLE"}}
	  	{"ip":"8.8.8.8"} >>> {"ip":"8.8.8.8","geo":{"continent":"North America","country":"United States","latitude":37.751,"longitude":-97.822,"accuracy_radius":1000,"timezone":"America/Chicago"}}

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "ip_database",
		"settings": {
			"options": {
				"type": "maxmind_geo",
				"settings": {
					"database": "location://path/to/maxmind.mmdb"
				}
			},
			"input_key": "ip",
			"output_key": "geo"
		}
	}
*/
type IPDatabase struct {
	/*
		Options is a configuration passed directly to the internal IP database package. Similar to processors, each database has its own config requirements. See internal/ip/database for more information.

		Each database is lazy loaded on first invocation and can be loaded from a path on local disk, an HTTP(S) URL, or an AWS S3 URL.
	*/
	Options   config.Config    `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
	// IgnoreClose overrides attempts to close the processor.
	IgnoreClose bool `json:"ignore_close"`
}

// Close closes enrichment database resources opened by the IPDatabase processor.
func (p IPDatabase) Close(ctx context.Context) error {
	if p.IgnoreClose {
		return nil
	}

	db, err := ipdb.Factory(p.Options)
	if err != nil {
		return fmt.Errorf("close ip_database: %v", err)
	}

	if db.IsEnabled() {
		if err := db.Close(); err != nil {
			return fmt.Errorf("close ip_database: %v", err)
		}
	}

	return nil
}

// ApplyBatch processes a slice of encapsulated data with the IPDatabase processor. Conditions are optionally applied to the data to enable processing.
func (p IPDatabase) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process ip_database: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process ip_database: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the IPDatabase processor.
func (p IPDatabase) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return capsule, fmt.Errorf("process ip_database: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
	}

	db, err := ipdb.Factory(p.Options)
	if err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	// lazy load the database
	if !db.IsEnabled() {
		if err := db.Open(ctx); err != nil {
			return capsule, fmt.Errorf("process ip_database: %v", err)
		}
	}

	res := capsule.Get(p.InputKey).String()
	record, err := db.Get(res)
	if err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	if err := capsule.Set(p.OutputKey, record); err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	return capsule, nil
}
