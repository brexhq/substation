package ip

import (
	"context"
	"fmt"
	"os"

	"github.com/brexhq/substation/internal/file"
	"github.com/ip2location/ip2location-go/v9"
)

type IP2Location struct {
	db *ip2location.DB
}

func (i *IP2Location) IsEnabled() bool {
	return i.db != nil
}

func (i *IP2Location) Load(ctx context.Context) error {
	location, exists := os.LookupEnv("IP2LOCATION_DB")
	if !exists {
		return fmt.Errorf("ip db %s: location not found", "IP2LOCATION_DB")
	}

	path, err := file.Get(ctx, location)
	defer os.Remove(path)

	if err != nil {
		return err
	}

	if i.db, err = ip2location.OpenDB(path); err != nil {
		return err
	}

	return nil
}

func (i *IP2Location) Location(ip string) (*location, error) {
	resp, err := i.db.Get_all(ip)
	if err != nil {
		return nil, err
	}

	loc := &location{
		Country:   resp.Country_long,
		City:      resp.City,
		Region:    resp.Region,
		Latitude:  float64(resp.Latitude),
		Longitude: float64(resp.Longitude),
	}

	return loc, nil
}
