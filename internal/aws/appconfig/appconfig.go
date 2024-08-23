// package appconfig provides functions for interacting with AWS AppConfig.
package appconfig

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/brexhq/substation/v2/internal/http"
)

// errMissingPrefetchEnvVar is returned when a Lambda is deployed without a configured AppConfig URL.
var errMissingPrefetchEnvVar = fmt.Errorf("missing AWS_APPCONFIG_EXTENSION_PREFETCH_LIST environment variable")

var client http.HTTP

// GetPrefetch queries and returns the Lambda's prefetched AppConfig configuration.
func GetPrefetch(ctx context.Context, dst io.Writer) error {
	if !client.IsEnabled() {
		client.Setup()
	}

	env := "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST"
	url, found := os.LookupEnv(env)
	if !found {
		return fmt.Errorf("appconfig getprefetch: %v", errMissingPrefetchEnvVar)
	}

	local := "http://localhost:2772" + url

	ctx = context.WithoutCancel(ctx)
	resp, err := client.Get(ctx, local)
	if err != nil {
		return fmt.Errorf("appconfig getprefetch URL %s: %v", local, err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(dst, resp.Body); err != nil {
		return fmt.Errorf("appconfig getprefetch: %v", err)
	}

	return nil
}
