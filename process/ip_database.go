//go:build !wasm

package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
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
type procIPDatabase struct {
	process
	Options config.Config `json:"options"`

	db ipdb.OpenCloser
}

// String returns the processor settings as an object.
func (p procIPDatabase) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procIPDatabase) Close(ctx context.Context) error {
	if p.IgnoreClose {
		return nil
	}

	db, err := ipdb.Get(p.Options)
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

// Create a new IP database processor.
func newProcIPDatabase(cfg config.Config) (p procIPDatabase, err error) {
	err = config.Decode(cfg.Settings, &p)
	if err != nil {
		return procIPDatabase{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
	if err != nil {
		return procIPDatabase{}, err
	}

	// only supports JSON, fail if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return procIPDatabase{}, fmt.Errorf("process: ip_database: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	p.db, err = ipdb.Get(p.Options)
	if err != nil {
		return procIPDatabase{}, fmt.Errorf("process: ip_database: %v", err)
	}

	if !p.db.IsEnabled() {
		if err := p.db.Open(context.Background()); err != nil {
			return procIPDatabase{}, fmt.Errorf("process: ip_database: %v", err)
		}
	}

	return p, nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procIPDatabase) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procIPDatabase) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	res := capsule.Get(p.Key).String()
	record, err := p.db.Get(res)
	if err != nil {
		return capsule, fmt.Errorf("process: ip_database: %v", err)
	}

	if err := capsule.Set(p.SetKey, record); err != nil {
		return capsule, fmt.Errorf("process: ip_database: %v", err)
	}

	return capsule, nil
}
