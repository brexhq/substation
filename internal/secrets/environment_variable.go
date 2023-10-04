package secrets

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type environmentVariableConfig struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TTLOffset int64  `json:"ttl_offset"`
}

func (c *environmentVariableConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *environmentVariableConfig) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("id: %v", errors.ErrMissingRequiredOption)
	}

	if c.Name == "" {
		return fmt.Errorf("name: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type env struct {
	conf environmentVariableConfig

	ttl int64
}

func newEnvironmentVariable(_ context.Context, cfg config.Config) (*env, error) {
	conf := environmentVariableConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("secrets: environment_variable: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("secrets: environment_variable: %v", err)
	}

	ttl := conf.TTLOffset
	if ttl == 0 {
		ttl = defaultTTL
	}

	return &env{
		conf: conf,
		ttl:  time.Now().Add(time.Duration(ttl) * time.Second).Unix(),
	}, nil
}

func (c *env) Collect(ctx context.Context) error {
	if v, ok := os.LookupEnv(c.conf.Name); ok {
		// SetWithTTL isn't used here because the TTL is managed by
		// transform/utility_secret.go.
		if err := cache.Set(ctx, c.conf.ID, v); err != nil {
			return fmt.Errorf("secrets: environment_variable: id %s: %v", c.conf.ID, err)
		}
	}

	return nil
}

func (c *env) Expired() bool {
	return time.Now().Unix() >= c.ttl
}
