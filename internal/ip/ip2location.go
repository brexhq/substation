package ip

import (
	"context"
	"os"

	"github.com/brexhq/substation/internal/file"
	"github.com/ip2location/ip2location-go/v9"
)

// IP2Location provides read access to an IP2Location binary database.
type IP2Location struct {
	db *ip2location.DB
}

// IsEnabled returns true if the database is open and ready for use.
func (i *IP2Location) IsEnabled() bool {
	return i.db != nil
}

// Close closes the open database.
func (i *IP2Location) Close() error {
	if i.IsEnabled() {
		i.db.Close()
	}

	return nil
}

// Setup contextually retrieves and opens an IP2Location BIN database. The location of the database is retrieved from the IP2LOCATION_DB environment variable and can be read from local disk, HTTP(S) URL, or AWS S3 URL. If the environment variable is missing, then there is no attempt to load the database.
func (i *IP2Location) Setup(ctx context.Context) error {
	db, ok := os.LookupEnv("IP2LOCATION_DB")
	if ok && !i.IsEnabled() {
		path, err := file.Get(ctx, db)
		defer os.Remove(path)

		if err != nil {
			return err
		}

		if i.db, err = ip2location.OpenDB(path); err != nil {
			return err
		}
	}

	return nil
}

// Location returns geolocation information for an IP address from an IP2Location BIN database.
func (i *IP2Location) Location(ip string) (*location, error) {
	resp, err := i.db.Get_all(ip)
	if err != nil {
		return nil, err
	}

	loc := &location{
		Country:   resp.Country_long,
		City:      resp.City,
		Region:    resp.Region,
		Latitude:  resp.Latitude,
		Longitude: resp.Longitude,
	}

	return loc, nil
}
