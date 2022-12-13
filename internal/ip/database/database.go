// package database provides tools for enriching IP addresses from enrichment databases.
package database

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/ip"
)

// errInvalidFactoryInput is returned when an unsupported OpenCloser is referenced in Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

// OpenCloser provides tools for opening, closing, and getting values from IP address enrichment databases.
type OpenCloser interface {
	ip.Getter
	Open(context.Context) error
	Close() error
	IsEnabled() bool
}

// Factory returns an OpenCloser. The returned OpenCloser must be opened before it can be used.
// func Factory(db string) (OpenCloser, error) {
func Factory(cfg config.Config) (OpenCloser, error) {
	switch t := cfg.Type; t {
	case "ip2location":
		var db IP2Location
		_ = config.Decode(cfg.Settings, &db)
		return &db, nil
	case "maxmind_asn":
		var db MaxMindASN
		_ = config.Decode(cfg.Settings, &db)
		return &db, nil
	case "maxmind_city":
		var db MaxMindCity
		_ = config.Decode(cfg.Settings, &db)
		return &db, nil
	default:
		return nil, fmt.Errorf("database %s: %v", t, errInvalidFactoryInput)
	}
}
