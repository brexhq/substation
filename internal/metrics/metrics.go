package metrics

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

const (
	metricsApplication = "Substation"
)

// Data contains a metric that can be sent to external services.
type Data struct {
	// Contextual information related to the metric. If the external service accepts key-value pairs (e.g., identifiers, tags), then this is passed directly to the service.
	Attributes map[string]string

	// A short name that describes the metric. This is passed directly to the external service and should use the upper camel case (UpperCamelCase) naming convention.
	Name string

	// The metric data point. This value is converted to the correct data type before being sent to the external service.
	Value interface{}
}

// AddAttributes is a convenience method for adding attributes to a metric.
func (d *Data) AddAttributes(attr map[string]string) {
	if d.Attributes == nil {
		d.Attributes = make(map[string]string)
	}

	for key, val := range attr {
		d.Attributes[key] = val
	}
}

type generator interface {
	Generate(context.Context, Data) error
}

func New(ctx context.Context, cfg config.Config) (generator, error) {
	switch cfg.Type {
	case "aws_cloudwatch_embedded_metrics":
		return newAWSCloudWatchEmbeddedMetrics(ctx, cfg)
	default:
		return nil, fmt.Errorf("metrics: new: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}
