package sink

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/jshlbrd/go-aggregate"

	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

/*
SumoLogic sinks JSON data to Sumo Logic using an HTTP collector. More information about Sumo Logic HTTP collectors is available here: https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Data-to-an-HTTP-Source.

The sink has these settings:
	URL:
		HTTP(S) endpoint that data is sent to
	CategoryKey (optional):
		JSON key-value that is used as the Sumo Logic source category
		defaults to no source category, which sends data to the source category configured for URL

The sink uses this Jsonnet configuration:
	{
		type: 'sumologic',
		settings: {
			url: 'foo.com/bar',
			category_key: 'foo',
		},
	}
*/
type SumoLogic struct {
	URL            string `json:"url"`
	CategoryKey    string `json:"category_key"`
	ErrorOnFailure bool   `json:"error_on_failure"`
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

	bundles := map[string]*aggregate.Bytes{}

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

			category := json.Get(data, sink.CategoryKey).String()
			if _, ok := bundles[category]; !ok {
				bundles[category] = &aggregate.Bytes{}
				bundles[category].New(1000*1000*.9, 10000)
			}

			// add event data to the category bundle
			// if category bundle is full, then send the bundle
			ok, err := bundles[category].Add(data)
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
				bundle := bundles[category].Get()
				for _, b := range bundle {
					buf.WriteString(fmt.Sprintf("%s\n", b))
				}

				if _, err := sumoLogicClient.Post(ctx, sink.URL, buf.Bytes(), h...); err != nil {
					return fmt.Errorf("err failed to POST to URL %s: %v", sink.URL, err)
				}

				log.WithField(
					"category", category,
				).WithField(
					"count", bundles[category].Count(),
				).Debug("sent events to Sumo Logic")

				bundles[category].Reset()
				bundles[category].Add(data)
			}
		}
	}

	// iterate and send remaining category bundles
	for category := range bundles {
		count := bundles[category].Count()
		if count == 0 {
			continue
		}

		h := headers
		h = append(h, http.Header{
			Key:   "X-Sumo-Category",
			Value: category,
		})

		var buf bytes.Buffer
		bundle := bundles[category].Get()
		for _, b := range bundle {
			buf.WriteString(fmt.Sprintf("%s\n", b))
		}

		if _, err := sumoLogicClient.Post(ctx, sink.URL, buf.Bytes(), h...); err != nil {
			return fmt.Errorf("err failed to POST to URL %s: %v", sink.URL, err)
		}

		log.WithField(
			"category", category,
		).WithField(
			"count", count,
		).Debug("sent events to Sumo Logic")
	}

	return nil
}
