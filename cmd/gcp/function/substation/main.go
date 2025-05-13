package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"

	"github.com/brexhq/substation/v2"

	"github.com/brexhq/substation/v2/internal/file"
)

// errFunctionMissingHandler is returned when the Function is deployed without a configured handler.
var errFunctionMissingHandler = fmt.Errorf("SUBSTATION_FUNCTION_HANDLER environment variable is missing")

func init() {
	handler, ok := os.LookupEnv("SUBSTATION_FUNCTION_HANDLER")
	if !ok {
		panic(fmt.Errorf("init handler %s: %v", handler, errFunctionMissingHandler))
	}

	switch handler {
	case "GCP_STORAGE":
		funcframework.RegisterCloudEventFunctionContext(context.Background(), "/", cloudStorageHandler)
	default:
		panic(fmt.Errorf("init handler %s: %v", handler, errFunctionMissingHandler))
	}
}

type customConfig struct {
	substation.Config

	Concurrency int `json:"concurrency"`
}

func getConfig(ctx context.Context) (io.Reader, error) {
	buf := new(bytes.Buffer)

	cfg, found := os.LookupEnv("SUBSTATION_CONFIG")
	if !found {
		return nil, fmt.Errorf("no config found")
	}

	path, err := file.Get(ctx, cfg)
	defer os.Remove(path)

	if err != nil {
		return nil, err
	}

	conf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer conf.Close()

	if _, err := io.Copy(buf, conf); err != nil {
		return nil, err
	}

	return buf, nil
}

func main() {
	// Use PORT environment variable, or default to 8080
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
