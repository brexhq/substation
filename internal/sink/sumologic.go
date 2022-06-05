package sink

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/jshlbrd/go-aggregate"

	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

// SumoLogicSinkInvalidJSON is returned when the Sumo Logic sink receives invalid JSON. If this error occurs, then parse the data into valid JSON or drop invalid JSON before it reaches the sink.
const SumoLogicSinkInvalidJSON = errors.Error("SumoLogicSinkInvalidJSON")

/*
SumoLogic sinks JSON data to Sumo Logic using an HTTP collector. More information about Sumo Logic HTTP collectors is available here: https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Data-to-an-HTTP-Source.

The sink has these settings:
	URL:
		HTTP(S) endpoint that data is sent to
	Category (optional):
		configured Sumo Logic source category
		defaults to no source category, which sends data to the source category configured for URL
	CategoryKey (optional):
		JSON key-value that is used as the Sumo Logic source category, overrides Category
		defaults to no source category, which sends data to the source category configured for URL

The sink uses this Jsonnet configuration:
	{
		type: 'sumologic',
		settings: {
			url: 'foo.com/bar',
			category: 'foo',
		},
	}
*/
type SumoLogic struct {
	URL         string `json:"url"`
	Category    string `json:"category"`
	CategoryKey string `json:"category_key"`
}

var sumoLogicClient http.HTTP

// Send sinks a channel of bytes with the SumoLogic sink.
func (sink *SumoLogic) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
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

	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			// Sumo Logic only parses JSON
			// if this error occurs, then parse the data into JSON
			if !json.Valid(data) {
				log.Info("Sumo Logic sink received invalid JSON data")
				return SumoLogicSinkInvalidJSON
			}

			if sink.CategoryKey != "" {
				category = json.Get(data, sink.CategoryKey).String()
			}

			if _, ok := buffer[category]; !ok {
				// aggregate up to 0.9MB or 10,000 items
				// https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source#Data_payload_considerations
				buffer[category] = &aggregate.Bytes{}
				buffer[category].New(1000*1000*.9, 10000)
			}

			// add data to the buffer
			// if buffer is full, then send the aggregated data
			ok, err := buffer[category].Add(data)
			if err != nil {
				return err
			}

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
					return fmt.Errorf("err failed to POST to URL %s: %v", sink.URL, err)
				}

				log.WithField(
					"category", category,
				).WithField(
					"count", buffer[category].Count(),
				).Debug("sent events to Sumo Logic")

				buffer[category].Reset()
				buffer[category].Add(data)
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
			return fmt.Errorf("err failed to POST to URL %s: %v", sink.URL, err)
		}

		log.WithField(
			"count", count,
		).WithField(
			"category", category,
		).Debug("sent events to Sumo Logic")
	}

	return nil
}
