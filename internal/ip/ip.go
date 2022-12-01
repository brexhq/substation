// package ip provides tools for modifying and enriching IP address data.
package ip

type location struct {
	Continent      string  `json:"continent,omitempty"`
	Country        string  `json:"country,omitempty"`
	City           string  `json:"city,omitempty"`
	Region         string  `json:"region,omitempty"`
	PostalCode     string  `json:"postal_code,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	AccuracyRadius uint16  `json:"accuracy_radius,omitempty"`
	Timezone       string  `json:"timezone,omitempty"`
}

type as struct {
	Number       uint   `json:"number,omitempty"`
	Organization string `json:"organization,omitempty"`
}
