// package ip provides tools for modifying IP address data.
package ip

// EnrichmentRecord is an aggregation of information commonly provided by IP address enrichment services.
type EnrichmentRecord struct {
	ASN      *ASN      `json:"asn,omitempty"`
	Location *Location `json:"location,omitempty"`
}

// Location is an abstracted data structure used for returning geolocation enrichment results.
type Location struct {
	Coordinates Coordinates `json:"coordinates,omitempty"`
	Continent   string      `json:"continent,omitempty"`
	Country     string      `json:"country,omitempty"`
	Region      string      `json:"region,omitempty"`
	City        string      `json:"city,omitempty"`
	PostalCode  string      `json:"postal_code,omitempty"`
	TimeZone    string      `json:"time_zone,omitempty"`
	Accuracy    float32     `json:"accuracy,omitempty"`
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
