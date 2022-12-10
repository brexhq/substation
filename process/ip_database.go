package process

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	ipdb "github.com/brexhq/substation/internal/ip/database"
)

var ipDatabases = make(map[string]ipdb.OpenCloser)

/*
IPDatabase processes data by querying IP addresses in enrichment databases, including geographic location (geo) and autonomous system (asn) databases. The processor supports multiple database providers by contextually retrieving and loading databases using environment variables and can be reused if multiple databases need to be queried.

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
				"function": "maxmind_geo"
			},
			"input_key": "ip",
			"output_key": "geo"
		}
	}
*/
type IPDatabase struct {
	Options   IPDatabaseOptions `json:"options"`
	Condition condition.Config  `json:"condition"`
	InputKey  string            `json:"input_key"`
	OutputKey string            `json:"output_key"`
}

/*
IPDatabaseOptions contains custom options for the IPDatabase processor.

	Function:
		Selects the enrichment database queried by the processor.

		The database is contextually retrieved using an environment variable and lazy loaded on first invocation. Each environment variable should contain the location of the database, which can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL.

		Must be one of:
			ip2location (IP2LOCATION)
			maxmind_asn (MAXMIND_ASN)
			maxmind_city (MAXMIND_CITY)
*/
type IPDatabaseOptions struct {
	Function string `json:"function"`
}

// Close closes enrichment database resources opened by the IPDatabase processor.
func (p IPDatabase) Close(ctx context.Context) error {
	for _, db := range ipDatabases {
		if err := db.Close(); err != nil {
			return fmt.Errorf("process ip_database: %v", err)
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

	// lazy load IP enrichment database
	// location of the database is set by environment variable
	// db must go into the map after opening to avoid race conditions
	if _, ok := ipDatabases[p.Options.Function]; !ok {
		location := os.Getenv(strings.ToUpper(p.Options.Function))

		db, err := ipdb.Factory(p.Options.Function)
		if err != nil {
			return capsule, fmt.Errorf("process ip_database: %v", err)
		}

		if err := db.Open(ctx, location); err != nil {
			return capsule, fmt.Errorf("process ip_database: %v", err)
		}

		ipDatabases[p.Options.Function] = db
	}

	res := capsule.Get(p.InputKey).String()
	record, err := ipDatabases[p.Options.Function].Get(res)
	if err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	if err := capsule.Set(p.OutputKey, record); err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	return capsule, nil
}
