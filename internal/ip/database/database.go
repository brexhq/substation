// package database provides tools for enriching IP addresses from enrichment databases.
package database

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/ip"
)

// databases are global variables that can be accessed across the application by using the Factory function.
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

// Get returns a pointer to an OpenCloser that is stored as a package level global variable. The OpenCloser must be opened before it can be used.
func Get(cfg config.Config) (OpenCloser, error) {
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
		return nil, fmt.Errorf("database %s: %v", t, errors.ErrInvalidFactoryInput)
	}
}
