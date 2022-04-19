package sink

import (
	"context"
	"fmt"
	"os"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

/*
SumoLogic implements the Sink interface and POSTs data to a Sumo Logic HTTP source. More information is available in the README.

URL: HTTP(S) endpoint that data is sent to, this is defined in the Sumo Logic console
Categories: array of options for inspecting input data and assigning a Sumo Logic source category
Categories.Condition: conditions that must pass to assign a Sumo Logic source category
Categories.Category: the Sumo Logic source category that is assigned
ErrorOnFailure (optional): determines if invalid input data causes the sink to error; defaults to false
*/
type SumoLogic struct {
	client     http.HTTP
	URL        string `mapstructure:"url"`
	Categories []struct {
		Condition condition.OperatorConfig `mapstructure:"condition"`
		Category  string                   `mapstructure:"category"`
	} `mapstructure:"categories"`
	ErrorOnFailure bool `mapstructure:"error_on_failure"`
}

// Send sends a channel of bytes to the Sumo Logic HTTP source categories defined by this sink.
func (sink *SumoLogic) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !sink.client.IsEnabled() {
		sink.client.Setup()
		if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
			sink.client.EnableXRay()
		}
	}

	bundles := map[string]*http.Aggregate{}

	headers := []http.Header{
		{
			Key:   "Content-Type",
			Value: "application/json",
		},
	}

	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			// Sumo Logic only parses JSON
			// if this error occurs, then parse the data into JSON
			if !json.Valid(data) && sink.ErrorOnFailure {
				return fmt.Errorf("err Sumo Logic sink received invalid JSON data: %v", json.JSONInvalidData)
			} else if !json.Valid(data) {
				log.Info("Sumo Logic sink received invalid JSON data")
				continue
			}

			var category string

			for _, categories := range sink.Categories {
				op, err := condition.OperatorFactory(categories.Condition)
				if err != nil {
					return err
				}
				ok, err := op.Operate(data)
				if err != nil {
					return err
				}

				if !ok {
					continue
				}

				category = categories.Category
			}

			if _, ok := bundles[category]; !ok {
				bundles[category] = &http.Aggregate{}
			}

			dataString := string(data)
			// add event data to the category bundle
			// if category bundle is full, then send the bundle
			ok := bundles[category].Add(dataString)
			if !ok {
				h := headers
				h = append(h, http.Header{
					Key:   "X-Sumo-Category",
					Value: category,
				})

				data := bundles[category].Get()
				if _, err := sink.client.Post(ctx, sink.URL, data, h...); err != nil {
					return fmt.Errorf("err failed to POST to URL %s: %v", sink.URL, err)
				}

				log.WithField(
					"category", category,
				).WithField(
					"count", bundles[category].Count(),
				).Debug("sent events to Sumo Logic")

				bundles[category] = &http.Aggregate{}
				bundles[category].Add(dataString)
			}
		}
	}

	// iterate and send remaining category bundles
	for category := range bundles {
		agg := bundles[category]
		size := agg.Size()
		if size == 0 {
			continue
		}

		h := headers
		h = append(h, http.Header{
			Key:   "X-Sumo-Category",
			Value: category,
		})

		data := agg.Get()
		if _, err := sink.client.Post(ctx, sink.URL, data, h...); err != nil {
			return fmt.Errorf("err failed to POST to URL %s: %v", sink.URL, err)
		}

		log.WithField(
			"category", category,
		).WithField(
			"count", agg.Count(),
		).Debug("sent events to Sumo Logic")
	}

	return nil
}
