package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	ipdb "github.com/brexhq/substation/internal/ip/database"
)

// ipDatabase processes data by querying IP addresses in enrichment databases, including
// geographic location (geo) and autonomous system (asn) databases. The processor supports
// multiple database providers and can be reused if multiple databases need to be queried.
// IP address information is abstracted from each enrichment database into a single record
// that contains these categories:
//
// - asn (autonomous system information)
//
// - geo (location information)
//
// See internal/ip/database for information on supported database providers.
//
// This processor supports the object handling pattern.
type _ipDatabase struct {
	process
	Options config.Config `json:"options"`
}

// String returns the processor settings as an object.
func (p _ipDatabase) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _ipDatabase) Close(ctx context.Context) error {
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

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _ipDatabase) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p _ipDatabase) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process ip_database: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
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

	res := capsule.Get(p.Key).String()
	record, err := db.Get(res)
	if err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	if err := capsule.Set(p.SetKey, record); err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	return capsule, nil
}
