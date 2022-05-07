package appconfig

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
)

// LambdaMissingAppConfig is returned when a Lambda is deployed without a configured AppConfig URL.
const LambdaMissingAppConfig = errors.Error("LambdaMissingAppConfig")

var client http.HTTP

// GetPrefetch queries and returns the Lambda's prefetched AppConfig configuration.
func GetPrefetch(ctx context.Context) ([]byte, error) {
	if !client.IsEnabled() {
		client.Setup()
	}

	url, found := os.LookupEnv("AWS_APPCONFIG_EXTENSION_PREFETCH_LIST")
	if !found {
		return nil, LambdaMissingAppConfig
	}

	resp, err := client.Get(ctx, "http://localhost:2772"+url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
