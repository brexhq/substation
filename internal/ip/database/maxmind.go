//go:build !wasm

package database

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/internal/ip"
	"github.com/oschwald/geoip2-golang"
)

// MaxMindASN provides read access to a MaxMind ASN database. The database is safe for concurrent access.
type MaxMindASN struct {
	// Database contains the location of the MaxMind City database. This can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL.
	Database string `json:"database"`
	/*
		Language determines the language that localized name data is returned as. More information is available here: https://support.maxmind.com/hc/en-us/articles/4414877149467-IP-Geolocation-Data.

		This is optional and defaults to "en" (English).
	*/
	Language string `json:"language"`
	mu       sync.RWMutex
	reader   *geoip2.Reader
}

// IsEnabled returns true if the database is open and ready for use.
func (d *MaxMindASN) IsEnabled() bool {
	return d.reader != nil
}

// Open retrieves the database and opens it for querying.
func (d *MaxMindASN) Open(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// avoids unnecessary opening
	if d.reader != nil {
		return nil
	}

	if d.Language == "" {
		d.Language = "en"
	}

	path, err := file.Get(ctx, d.Database)
	defer os.Remove(path)

	if err != nil {
		return fmt.Errorf("database: %v", err)
	}

	if d.reader, err = geoip2.Open(path); err != nil {
		return fmt.Errorf("database: %v", err)
	}

	return nil
}

// Close closes the open database.
func (d *MaxMindASN) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// avoids unnecessary closing
	if d.reader == nil {
		return nil
	}

	if err := d.reader.Close(); err != nil {
		return fmt.Errorf("database: %v", err)
	}

	// reader is made nil so that IsEnabled correctly
	// returns the non-enabled state
	d.reader = nil
	return nil
}

// Get queries the database and returns an aggregated database record containing enrichment information.
func (d *MaxMindASN) Get(addr string) (*ip.EnrichmentRecord, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	paddr := net.ParseIP(addr)
	if paddr == nil {
		return nil, fmt.Errorf("database: %v", ip.ErrInvalidIPAddress)
	}

	resp, err := d.reader.ASN(paddr)
	if err != nil {
		return nil, err
	}

	rec := &ip.EnrichmentRecord{
		ASN: &ip.ASN{
			Number:       resp.AutonomousSystemNumber,
			Organization: resp.AutonomousSystemOrganization,
		},
	}

	return rec, nil
}

// MaxMindCity provides read access to a MaxMind City database. The database is safe for concurrent access.
type MaxMindCity struct {
	// Database contains the location of the MaxMind City database. This can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL.
	Database string `json:"database"`
	/*
		Language determines the language that localized name data is returned as. More information is available here: https://support.maxmind.com/hc/en-us/articles/4414877149467-IP-Geolocation-Data.

		This is optional and defaults to "en" (English).
	*/
	Language string `json:"language"`
	mu       sync.RWMutex
	reader   *geoip2.Reader
}

// IsEnabled returns true if the database is open and ready for use.
func (d *MaxMindCity) IsEnabled() bool {
	return d.reader != nil
}

// Open retrieves the database and opens it for querying.
func (d *MaxMindCity) Open(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// avoids unnecessary opening
	if d.reader != nil {
		return nil
	}

	if d.Language == "" {
		d.Language = "en"
	}

	path, err := file.Get(ctx, d.Database)
	defer os.Remove(path)

	if err != nil {
		return fmt.Errorf("database: %v", err)
	}

	if d.reader, err = geoip2.Open(path); err != nil {
		return fmt.Errorf("database: %v", err)
	}

	return nil
}

// Close closes the open database.
func (d *MaxMindCity) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// avoids unnecessary closing
	if d.reader == nil {
		return nil
	}

	if err := d.reader.Close(); err != nil {
		return fmt.Errorf("database: %v", err)
	}

	// reader is made nil so that IsEnabled correctly
	// returns the non-enabled state
	d.reader = nil
	return nil
}

// Get queries the database and returns an aggregated database record containing enrichment information.
func (d *MaxMindCity) Get(addr string) (*ip.EnrichmentRecord, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	paddr := net.ParseIP(addr)
	if paddr == nil {
		return nil, fmt.Errorf("database: %v", ip.ErrInvalidIPAddress)
	}

	resp, err := d.reader.City(paddr)
	if err != nil {
		return nil, fmt.Errorf("database: %v", err)
	}

	rec := &ip.EnrichmentRecord{
		Location: &ip.Location{
			Coordinates: &ip.Coordinates{
				Latitude:  float32(resp.Location.Latitude),
				Longitude: float32(resp.Location.Longitude),
			},
			Continent:  resp.Continent.Names[d.Language],
			Country:    resp.Country.Names[d.Language],
			City:       resp.City.Names[d.Language],
			PostalCode: resp.Postal.Code,
			Accuracy:   float32(resp.Location.AccuracyRadius),
			TimeZone:   resp.Location.TimeZone,
		},
	}

	return rec, nil
}
