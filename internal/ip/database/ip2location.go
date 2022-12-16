package database

import (
	"context"
	"os"
	"sync"

	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/internal/ip"
	"github.com/ip2location/ip2location-go/v9"
)

// IP2Location provides read access to an IP2Location binary database.
type IP2Location struct {
	Database string `json:"database"`
	mu       sync.RWMutex
	db       *ip2location.DB
}

// IsEnabled returns true if the database is open and ready for use.
func (d *IP2Location) IsEnabled() bool {
	return d.db != nil
}

// Open retrieves the database and opens it for querying. The location of the database can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL.
func (d *IP2Location) Open(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// avoids unnecessary opening
	if d.db != nil {
		return nil
	}

	path, err := file.Get(ctx, d.Database)
	defer os.Remove(path)

	if err != nil {
		return err
	}

	if d.db, err = ip2location.OpenDB(path); err != nil {
		return err
	}

	return nil
}

// Close closes the open database.
func (d *IP2Location) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// avoids unnecessary closing
	if d.db == nil {
		return nil
	}

	d.db.Close()

	// db is made nil so that IsEnabled correctly
	// returns the non-enabled state
	d.db = nil
	return nil
}

// Get queries the database and returns an aggregated database record containing enrichment information.
func (d *IP2Location) Get(addr string) (*ip.EnrichmentRecord, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	resp, err := d.db.Get_all(addr)
	if err != nil {
		return nil, err
	}

	db := &ip.EnrichmentRecord{
		Location: &ip.Location{
			Coordinates: &ip.Coordinates{
				Latitude:  resp.Latitude,
				Longitude: resp.Longitude,
			},
			Country:  resp.Country_long,
			City:     resp.City,
			Region:   resp.Region,
			TimeZone: resp.Timezone,
		},
	}

	return db, nil
}
