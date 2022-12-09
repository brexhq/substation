// package ip provides tools for modifying IP address data.
package ip

import "github.com/brexhq/substation/internal/errors"

// ErrInvalidIPAddress is returned when an invalid IP address is referenced in any function or method.
const ErrInvalidIPAddress = errors.Error("invalid IP address")

// Getter provides a method for getting an enrichment record from any IP address enrichment source.
type Getter interface {
	Get(string) (*EnrichmentRecord, error)
}

// EnrichmentRecord is an aggregation of information commonly provided by IP address enrichment services.
type EnrichmentRecord struct {
	ASN      *ASN      `json:"asn,omitempty"`
	Location *Location `json:"location,omitempty"`
}

// Location is an abstracted data structure used for returning geolocation enrichment results.
type Location struct {
	Coordinates *Coordinates `json:"coordinates,omitempty"`
	Continent   string       `json:"continent,omitempty"`
	Country     string       `json:"country,omitempty"`
	Region      string       `json:"region,omitempty"`
	City        string       `json:"city,omitempty"`
	PostalCode  string       `json:"postal_code,omitempty"`
	TimeZone    string       `json:"time_zone,omitempty"`
	Accuracy    float32      `json:"accuracy,omitempty"`
}

// Coordinates is an abstracted data structure used for returning coordinates enrichment results.
type Coordinates struct {
	Latitude  float32 `json:"latitude,omitempty"`
	Longitude float32 `json:"longitude,omitempty"`
}

// ASN is an abstracted data structure used for returning autonomous system number (ASN) enrichment results.
type ASN struct {
	Number       uint   `json:"number,omitempty"`
	Organization string `json:"organization,omitempty"`
}
