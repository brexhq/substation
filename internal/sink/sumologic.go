package sink

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/jshlbrd/go-aggregate"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

var sumoLogicClient http.HTTP

// errSumoLogicJSON is returned when the Sumo Logic sink receives invalid JSON. If this error occurs, then parse the data into valid JSON or drop invalid JSON before it reaches the sink.
const errSumoLogicJSON = errors.Error("input must be JSON")

// sumologic sinks objects to Sumo Logic using an HTTP collector.
//
// More information about Sumo Logic HTTP collectors is available here:
// https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Data-to-an-HTTP-Source.
type _sumologic struct {
	// URL is the Sumo Logic HTTPS endpoint that objects are sent to.
	URL string `json:"url"`
	// Category is the Sumo Logic source category that overrides the
	// configuration for the HTTPS endpoint.
	//
	// This is optional and has no default.
	Category string `json:"category"`
	// CategoryKey retrieves a value from an object that is used as
	// the Sumo Logic source category that overrides the configuration
	// for the HTTPS endpoint. If used, then this overrides Category.
	//
	// This is optional and has no default.
	CategoryKey string `json:"category_key"`
}

// Send sinks a channel of encapsulated data with the sink.
func (sink *_sumologic) Send(ctx context.Context, ch *config.Channel) error {
	if !sumoLogicClient.IsEnabled() {
		sumoLogicClient.Setup()
		if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
			sumoLogicClient.EnableXRay()
		}
	}

	buffer := map[string]*aggregate.Bytes{}

	headers := []http.Header{
		{
			Key:   "Content-Type",
			Value: "application/json",
		},
	}

	var category string
	if sink.Category != "" {
		category = sink.Category
	}

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if !json.Valid(capsule.Data()) {
				return fmt.Errorf("sink: sumologic category %s: %v", category, errSumoLogicJSON)
			}

			if sink.CategoryKey != "" {
				category = capsule.Get(sink.CategoryKey).String()
			}

			if _, ok := buffer[category]; !ok {
				// aggregate up to 0.9MB or 10,000 items
				// https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source#Data_payload_considerations
				buffer[category] = &aggregate.Bytes{}
				buffer[category].New(10000, 1000*1000*.9)
			}

			// add data to the buffer
			// if buffer is full, then send the aggregated data
			ok := buffer[category].Add(capsule.Data())
			if !ok {
				h := headers
				h = append(h, http.Header{
					Key:   "X-Sumo-Category",
					Value: category,
				})

				var buf bytes.Buffer
				items := buffer[category].Get()
				for _, i := range items {
					buf.WriteString(fmt.Sprintf("%s\n", i))
				}

				if _, err := sumoLogicClient.Post(ctx, sink.URL, buf.Bytes(), h...); err != nil {
					// Post err returns metadata
					return fmt.Errorf("sink: sumologic: %v", err)
				}

				log.WithField(
					"category", category,
				).WithField(
					"count", buffer[category].Count(),
				).Debug("sent events to Sumo Logic")

				buffer[category].Reset()
				_ = buffer[category].Add(capsule.Data())
			}
		}
	}

	// iterate and send remaining buffers
	for category := range buffer {
		count := buffer[category].Count()
		if count == 0 {
			continue
		}

		h := headers
		h = append(h, http.Header{
			Key:   "X-Sumo-Category",
			Value: category,
		})

		var buf bytes.Buffer
		bundle := buffer[category].Get()
		for _, b := range bundle {
			buf.WriteString(fmt.Sprintf("%s\n", b))
		}

		if _, err := sumoLogicClient.Post(ctx, sink.URL, buf.Bytes(), h...); err != nil {
			// Post err returns metadata
			return fmt.Errorf("sink: sumologic: %v", err)
		}

		log.WithField(
			"count", count,
		).WithField(
			"category", category,
		).Debug("sent events to Sumo Logic")
	}

	return nil
}
