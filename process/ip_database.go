package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	ipdb "github.com/brexhq/substation/internal/ip/database"
)

var ipDatabases = make(map[string]ipdb.OpenCloser)

type ipDatabase struct {
	process
	Options ipDatabaseOptions `json:"options"`
}

type ipDatabaseOptions struct {
	Function        string        `json:"function"`
	DatabaseOptions config.Config `json:"database_options"`
}

// Close closes enrichment database resources opened by the ipDatabase processor.
func (p ipDatabase) Close(ctx context.Context) error {
	for _, db := range ipDatabases {
		if err := db.Close(); err != nil {
			return fmt.Errorf("process ip_database: %v", err)
		}
	}

	return nil
}

func (p ipDatabase) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)

	if err != nil {
		return nil, fmt.Errorf("process ip_database: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the ipDatabase processor.
func (p ipDatabase) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process ip_database: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	// lazy load IP enrichment database
	// db must go into the map after opening to avoid race conditions
	if _, ok := ipDatabases[p.Options.Function]; !ok {
		db, err := ipdb.Factory(p.Options.DatabaseOptions)
		if err != nil {
			return capsule, fmt.Errorf("process ip_database: %v", err)
		}

		if err := db.Open(ctx); err != nil {
			return capsule, fmt.Errorf("process ip_database: %v", err)
		}

		ipDatabases[p.Options.Function] = db
	}

	res := capsule.Get(p.Key).String()
	record, err := ipDatabases[p.Options.Function].Get(res)
	if err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	if err := capsule.Set(p.SetKey, record); err != nil {
		return capsule, fmt.Errorf("process ip_database: %v", err)
	}

	return capsule, nil
}
