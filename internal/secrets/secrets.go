// Package secrets provides functions for retrieving local and remote secrets
package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/secretsmanager"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/kv"
	"github.com/brexhq/substation/internal/regexp"
)

var (
	// KV store is used as a secrets cache
	cache             kv.Storer
	secretsManagerAPI secretsmanager.API
)

// Regexp is used for interpolating secrets.
const Regexp = `{{(SECRETS_[A-Z]+:[^}]+)}}`

// errSecretNotFound is returned when Get is called but no secret is found.
const errSecretNotFound = errors.Error("secret not found")

// ttl enforces a 15 minute rotation for all secrets stored in memory.
const ttl = 15 * 60

/*
Get retrieves a secret from these locations (in order):

- Environment variables (SECRETS_ENV)

- AWS Secrets Manager (SECRETS_AWS)

Secret identification relies on the naming convention SECRETS_[LOCATION]:[POINTER]
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
			secretsManagerAPI.Setup()
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

// Interpolate identifies when a string contains a secret and interpolates
// the secret with the string.
func Interpolate(ctx context.Context, s string, exp string) (string, error) {
	if !strings.Contains(s, "{{SECRETS_") {
		return s, nil
	}

	re, err := regexp.Compile(exp)
	if err != nil {
		return "", err
	}

	match := re.FindStringSubmatch(s)
	if len(match) == 0 {
		return s, nil
	}

	secret, err := Get(ctx, match[len(match)-1])
	if err != nil {
		return "", err
	}

	return re.ReplaceAllString(s, secret), nil
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
