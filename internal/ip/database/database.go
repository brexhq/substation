// package database provides tools for enriching IP addresses from enrichment databases.
package database

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/ip"
)

// errInvalidFactoryInput is returned when an unsupported Databaser is referenced in Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

// Databaser provides tools for opening, managing, and reading enrichment information from IP address enrichment databases.
type Databaser interface {
	Read(string) (*ip.EnrichmentRecord, error)
	Open(context.Context, string) error
	IsEnabled() bool
	Close() error
}

// Factory returns a Databaser. The returned Databaser must be opened before it can be used.
func Factory(db string) (Databaser, error) {
	switch db {
	case "ip2location":
		return &IP2Location{}, nil
	case "maxmind_asn":
		return &MaxMindASN{}, nil
	case "maxmind_city":
		return &MaxMindCity{}, nil
	default:
		return nil, fmt.Errorf("ip database %s: %v", db, errInvalidFactoryInput)
	}
}
