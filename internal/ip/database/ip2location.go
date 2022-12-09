package database

import (
	"context"
	"os"

	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/internal/ip"
	"github.com/ip2location/ip2location-go/v9"
)

// IP2Location provides read access to an IP2Location binary database.
type IP2Location struct {
	db *ip2location.DB
}

// IsEnabled returns true if the database is open and ready for use.
func (d *IP2Location) IsEnabled() bool {
	return d.db != nil
}

// Open retrieves the database and opens it for querying. The location of the database can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL.
func (d *IP2Location) Open(ctx context.Context, location string) error {
	path, err := file.Get(ctx, location)
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
	if d.IsEnabled() {
		d.db.Close()
	}

	return nil
}

// Get queries the database and returns an aggregated database record containing enrichment information.
func (d *IP2Location) Get(addr string) (*ip.EnrichmentRecord, error) {
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
