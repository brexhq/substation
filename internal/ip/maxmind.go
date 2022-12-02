package ip

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/brexhq/substation/internal/file"
	"github.com/oschwald/geoip2-golang"
)

// MaxMind provides read access to MaxMind databases.
type MaxMind struct {
	asn      *geoip2.Reader
	geo      *geoip2.Reader
	language string
}

// IsASNEnabled returns true if the ASN database is open and ready for use.
func (m *MaxMind) IsASNEnabled() bool {
	return m.asn != nil
}

// IsLocationEnabled returns true if the location database is open and ready for use.
func (m *MaxMind) IsLocationEnabled() bool {
	return m.geo != nil
}

// Setup contextually retrieves and opens MaxMind databases. The location of each database is retrieved from the MAXMIND_ASN_DB and MAXMIND_LOCATION_DB environment variables and can be read from local disk, HTTP(S) URL, or AWS S3 URL. If an environment variable is missing, then there is no attempt to load the database.
func (m *MaxMind) Setup(ctx context.Context) error {
	m.SetLanguage()

	asn, ok := os.LookupEnv("MAXMIND_ASN_DB")
	if ok && !m.IsASNEnabled() {
		path, err := file.Get(ctx, asn)
		defer os.Remove(path)

		if err != nil {
			return err
		}

		if m.asn, err = geoip2.Open(path); err != nil {
			return err
		}
	}

	location, ok := os.LookupEnv("MAXMIND_LOCATION_DB")
	if ok && !m.IsLocationEnabled() {
		path, err := file.Get(ctx, location)
		defer os.Remove(path)

		if err != nil {
			return err
		}

		if m.geo, err = geoip2.Open(path); err != nil {
			return err
		}
	}

	return nil
}

// Close closes all open databases.
func (m *MaxMind) Close() error {
	if m.IsASNEnabled() {
		if err := m.asn.Close(); err != nil {
			return fmt.Errorf("ip: %v", err)
		}
	}

	if m.IsLocationEnabled() {
		if err := m.geo.Close(); err != nil {
			return fmt.Errorf("ip: %v", err)
		}
	}

	return nil
}

// SetLanguage configures the language that is used when reading values from MaxMind databases. The value is retrieved from the MAXMIND_LANGUAGE environment variable. If the environment variable is missing, then the default language is English.
func (m *MaxMind) SetLanguage() {
	lang, exists := os.LookupEnv("MAXMIND_LANGUAGE")
	if !exists {
		lang = "en"
	}
	m.language = lang
}

// ASN returns autonomous system information for an IP address from a MaxMind database.
func (m *MaxMind) ASN(ip string) (*asn, error) {
	pip := net.ParseIP(ip)
	resp, err := m.asn.ASN(pip)
	if err != nil {
		return nil, err
	}

	asn := &asn{
		Number:       resp.AutonomousSystemNumber,
		Organization: resp.AutonomousSystemOrganization,
	}

	return asn, nil
}

// Location returns geolocation information for an IP address from a MaxMind database.
func (m *MaxMind) Location(ip string) (*location, error) {
	pip := net.ParseIP(ip)
	resp, err := m.geo.City(pip)
	if err != nil {
		return nil, err
	}

	loc := &location{
		Continent:      resp.Continent.Names[m.language],
		Country:        resp.Country.Names[m.language],
		City:           resp.City.Names[m.language],
		PostalCode:     resp.Postal.Code,
		Latitude:       float32(resp.Location.Latitude),
		Longitude:      float32(resp.Location.Longitude),
		AccuracyRadius: resp.Location.AccuracyRadius,
		Timezone:       resp.Location.TimeZone,
	}

	if len(resp.Subdivisions) > 0 {
		loc.Region = resp.Subdivisions[0].Names[m.language]
	}

	return loc, nil
}
