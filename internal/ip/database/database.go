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

// databases are stored globally and can be accessed across the application by using the GlobalFactory function.
var (
	ip2loc      IP2Location
	maxMindASN  MaxMindASN
	maxMindCity MaxMindCity
)

// OpenCloser provides tools for opening, closing, and getting values from IP address enrichment databases.
type OpenCloser interface {
	ip.Getter
	Open(context.Context) error
	Close() error
	IsEnabled() bool
}

// Factory returns an OpenCloser. The OpenCloser must be opened before it can be used.
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

// GlobalFactory returns a pointer to an OpenCloser that is stored as a package level global variable. The OpenCloser must be opened before it can be used.
func GlobalFactory(cfg config.Config) (OpenCloser, error) {
	switch t := cfg.Type; t {
	case "ip2location":
		_ = config.Decode(cfg.Settings, &ip2loc)
		return &ip2loc, nil
	case "maxmind_asn":
		_ = config.Decode(cfg.Settings, &maxMindASN)
		return &maxMindASN, nil
	case "maxmind_city":
		_ = config.Decode(cfg.Settings, &maxMindCity)
		return &maxMindCity, nil
	default:
		return nil, fmt.Errorf("database %s: %v", t, errInvalidFactoryInput)
	}
}
