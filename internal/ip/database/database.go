// package database provides tools for enriching IP addresses from enrichment databases.
package database

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/errors"
)

// errInvalidFactoryInput is returned when an unsupported OpenCloser is referenced in Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

// OpenCloser provides tools for opening and closing IP address enrichment databases.
type OpenCloser interface {
	Open(context.Context, string) error
	Close() error
	IsEnabled() bool
}

// Factory returns an OpenCloser. The returned OpenCloser must be opened before it can be used.
func Factory(db string) (OpenCloser, error) {
	switch db {
	case "ip2location":
		return &IP2Location{}, nil
	case "maxmind_asn":
		return &MaxMindASN{}, nil
	case "maxmind_city":
		return &MaxMindCity{}, nil
	default:
		return nil, fmt.Errorf("database %s: %v", db, errInvalidFactoryInput)
	}
}
