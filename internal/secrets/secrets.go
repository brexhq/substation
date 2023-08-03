// Package secrets provides functions for retrieving local and remote secrets and interpolating them into configuration files
package secrets

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/brexhq/substation/config"
	_aws "github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/secretsmanager"
	"github.com/brexhq/substation/internal/kv"
)

var (
	// interpRe is used for parsing secrets during interpolation. Secrets
	// must not contain any curly braces.
	interpRe = regexp.MustCompile(`\${(SECRETS_[A-Z]+:[^}]+)}`)
	// KV store is used as a secrets cache
	cache             kv.Storer
	secretsManagerAPI secretsmanager.API
)

// errSecretNotFound is returned when Get is called but no secret is found.
var errSecretNotFound = fmt.Errorf("secret not found")

// ttl enforces a 15 minute rotation for all secrets stored in memory.
const ttl = 15 * 60

/*
Get retrieves a secret from these locations (in order):

- Environment variables (SECRETS_ENV)

- AWS Secrets Manager (SECRETS_AWS)

Secret identification relies on the naming convention SECRETS_[LOCATION]:[NAME]
so that select components of the system can identify and parse secrets. Secrets
should only be put into configurations and never in data or objects that flow
through the system. Not all components will use secrets; if the component does,
then it will self-identify its support in documentation.

If a secret is found, then it is stored in memory for up to 15 minutes; if no
secret is found, then errSecretNotFound is returned. This function is safe
for concurrent access.
*/
func Get(ctx context.Context, secret string) (string, error) {
	// use the cached secret if it exists
	// interactions with a KV store are always safe for concurrent access
	s, err := cache.Get(ctx, secret)
	if err != nil {
		return "", fmt.Errorf("secrets %s: %v", secret, err)
	}

	if s != nil {
		// secrets are always strings
		return s.(string), nil
	}

	// secrets kept in memory use a time-to-live to enforce rotation
	ttl := time.Now().Add(time.Duration(ttl) * time.Second).Unix()

	if strings.HasPrefix(secret, "SECRETS_ENV:") {
		r := strings.ReplaceAll(secret, "SECRETS_ENV:", "")
		if v, ok := os.LookupEnv(r); ok {
			if err := cache.SetWithTTL(ctx, secret, v, ttl); err != nil {
				return "", fmt.Errorf("secrets %s: %v", secret, err)
			}

			return v, nil
		}
	}

	if strings.HasPrefix(secret, "SECRETS_AWS:") {
		// AWS SDK client are always safe for concurrent access
		if !secretsManagerAPI.IsEnabled() {
			secretsManagerAPI.Setup(_aws.Config{})
		}

		r := strings.ReplaceAll(secret, "SECRETS_AWS:", "")
		v, err := secretsManagerAPI.GetSecret(ctx, r)
		if err != nil {
			return "", fmt.Errorf("secrets %s: %v", secret, err)
		}

		if err := cache.SetWithTTL(ctx, secret, v, ttl); err != nil {
			return "", fmt.Errorf("secrets %s: %v", secret, err)
		}

		return v, nil
	}

	return "", fmt.Errorf("secrets %s: %v", secret, errSecretNotFound)
}

// Interpolate identifies when a string contains one or more secrets and
// interpolates each secret with the string. This function uses the same
// convention as the standard library's regexp package for capturing named
// groups (${name}).
//
// For example, if the string is "/path/to/{SECRETS_ENV:FOO}" and BAR is the
// secret stored in the environment variable FOO, then the interpolated string
// is "/path/to/BAR".
//
// Multiple secrets can be stored in a single string; if the string is
// "/path/to/{SECRETS_ENV:FOO}/{SECRETS_ENV:BAZ}", then the interpolated string
// is "/path/to/BAR/QUX".
//
// If more than one interpolation function is applied to a string (e.g., non-secrets
// capture groups), then this function must be called first.
func Interpolate(ctx context.Context, s string) (string, error) {
	if !strings.Contains(s, "${SECRETS_") {
		return s, nil
	}

	matches := interpRe.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		if len(m) == 0 {
			continue
		}

		secret, err := Get(ctx, m[len(m)-1])
		if err != nil {
			return "", err
		}

		// replaces each substring with a secret.
		// if the secret is BAR and the string was
		// "/path/to/secret/{SECRETS_ENV:FOO}", then the interpolated
		// string output is "/path/to/secret/BAR".
		old := fmt.Sprintf("${%s}", m[len(m)-1])
		s = strings.Replace(s, old, secret, 1)
	}

	return s, nil
}

func init() {
	kv, err := kv.New(config.Config{
		Type: "memory",
		Settings: map[string]interface{}{
			// this can be converted to an env config if needed,
			// but otherwise the capacity seems fine
			"capacity": 100,
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
