package ip

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/brexhq/substation/internal/file"
	"github.com/oschwald/geoip2-golang"
)

type MaxMind struct {
	as       *geoip2.Reader
	geo      *geoip2.Reader
	language string
}

func (m *MaxMind) IsASEnabled() bool {
	return m.as != nil
}

func (m *MaxMind) IsGeoEnabled() bool {
	return m.geo != nil
}

func (m *MaxMind) LoadLanguage() {
	lang, exists := os.LookupEnv("MAXMIND_LANGUAGE")
	if !exists {
		lang = "en"
	}
	m.language = lang
}

// LoadGeo contextually retrieves and loads a MaxMind GeoIP2 or GeoLite2 database.
func (m *MaxMind) LoadGeo(ctx context.Context) error {
	m.LoadLanguage()

	location, exists := os.LookupEnv("MAXMIND_CITY_DB")
	if !exists {
		return fmt.Errorf("city location not found")
	}

	path, err := file.Get(ctx, location)
	defer os.Remove(path)

	if err != nil {
		return err
	}

	if m.geo, err = geoip2.Open(path); err != nil {
		return err
	}

	return nil
}

func (m *MaxMind) LoadAS(ctx context.Context) error {
	m.LoadLanguage()

	location, exists := os.LookupEnv("MAXMIND_ASN_DB")
	if !exists {
		return fmt.Errorf("asn location not found")
	}

	path, err := file.Get(ctx, location)
	defer os.Remove(path)

	if err != nil {
		return err
	}

	if m.as, err = geoip2.Open(path); err != nil {
		return err
	}

	return nil
}

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
		Latitude:       resp.Location.Latitude,
		Longitude:      resp.Location.Longitude,
		AccuracyRadius: resp.Location.AccuracyRadius,
		Timezone:       resp.Location.TimeZone,
	}

	if len(resp.Subdivisions) > 0 {
		loc.Region = resp.Subdivisions[0].Names[m.language]
	}

	return loc, nil
}

func (m *MaxMind) AS(ip string) (*as, error) {
	pip := net.ParseIP(ip)
	resp, err := m.as.ASN(pip)
	if err != nil {
		return nil, err
	}

	as := &as{
		Number:       resp.AutonomousSystemNumber,
		Organization: resp.AutonomousSystemOrganization,
	}

	return as, nil
}
