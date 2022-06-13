package appconfig

import (
	"context"
	"fmt"
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

	env := "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST"
	url, found := os.LookupEnv(env)
	if !found {
		return nil, fmt.Errorf("getprefetch lookup %s: %v", env, LambdaMissingAppConfig)
	}

	local := "http://localhost:2772" + url
	resp, err := client.Get(ctx, local)
	if err != nil {
		return nil, fmt.Errorf("getprefetch retrieve URL %s: %v", local, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("getprefetch read URL %s: %v", local, err)
	}

	return body, nil
}
