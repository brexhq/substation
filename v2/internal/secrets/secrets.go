// Package secrets provides functions for retrieving local and remote secrets and interpolating them into configuration files.
package secrets

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/internal/kv"
)

var (
	// interpRe is used for parsing secrets during interpolation. Secrets
	// must not contain any curly braces.
	interpRe = regexp.MustCompile(`\${(SECRET:[^}]+)}`)
	// KV store is used as a secrets cache
	cache kv.Storer
)

// defaultTTL enforces a 15 minute rotation for all secrets stored in memory.
const defaultTTL = "15m"

type Retriever interface {
	Retrieve(context.Context) error
	Expired() bool
}

func New(ctx context.Context, cfg config.Config) (Retriever, error) {
	switch cfg.Type {
	case "aws_secrets_manager":
		return newAWSSecretsManager(ctx, cfg)
	case "environment_variable":
		return newEnvironmentVariable(ctx, cfg)
	default:
		return nil, fmt.Errorf("secrets: new: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

// Interpolate identifies when a string contains one or more secrets and
// interpolates each secret with the string. This function uses the same
// convention as the standard library's regexp package for capturing named
// groups (${name}).
//
// For example, if the string is "/path/to/${SECRET:FOO}" and BAR is the
// secret value stored in the internal lookup, then the interpolated string
// is "/path/to/BAR".
//
// Multiple secrets can be stored in a single string; if the string is
// "/path/to/${SECRET:FOO}/${SECRET:BAZ}", then the interpolated string
// is "/path/to/BAR/QUX".
//
// If more than one interpolation function is applied to a string (e.g., non-secrets
// capture groups), then this function must be called first.
func Interpolate(ctx context.Context, s string) (string, error) {
	if !strings.Contains(s, "${SECRET") {
		return s, nil
	}

	matches := interpRe.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		if len(m) == 0 {
			continue
		}

		secretName := strings.ReplaceAll(m[len(m)-1], "SECRET:", "")
		secret, err := cache.Get(ctx, secretName)
		if err != nil {
			return "", err
		}

		// Replaces each substring with a secret. If the secret is
		// BAR and the string was "/path/to/secret/${SECRET:FOO}",
		// then the interpolated  string output is "/path/to/secret/BAR".
		old := fmt.Sprintf("${%s}", m[len(m)-1])
		s = strings.Replace(s, old, secret.(string), 1)
	}

	return s, nil
}

func init() {
	kv, err := kv.New(config.Config{
		Type: "memory",
		Settings: map[string]interface{}{
			"capacity": 1000,
		},
	})
	if err != nil {
		panic(err)
	}

	if err := kv.Setup(context.TODO()); err != nil {
		panic(err)
	}

	cache = kv
}
