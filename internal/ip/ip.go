// package ip provides tools for modifying and enriching IP address data.
package ip

// location is an abstracted data structure used for returning results from geolocation enrichment databases.
type location struct {
	Continent      string  `json:"continent,omitempty"`
	Country        string  `json:"country,omitempty"`
	City           string  `json:"city,omitempty"`
	Region         string  `json:"region,omitempty"`
	PostalCode     string  `json:"postal_code,omitempty"`
	Latitude       float32 `json:"latitude,omitempty"`
	Longitude      float32 `json:"longitude,omitempty"`
	AccuracyRadius uint16  `json:"accuracy_radius,omitempty"`
	Timezone       string  `json:"timezone,omitempty"`
}

// asn is an abstracted data structure used for returning results from ASN enrichment databases.
type asn struct {
	Number       uint   `json:"number,omitempty"`
	Organization string `json:"organization,omitempty"`
}
