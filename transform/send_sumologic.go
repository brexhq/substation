package transform

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/jshlbrd/go-aggregate"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

// errSendSumoLogicNonObject is returned when non-object data is sent to the transform.
//
// If this error occurs, then parse the data into an object (or drop invalid objects)
// before attempting to send the data.
var errSendSumoLogicNonObject = fmt.Errorf("input must be object")

type sendSumoLogicConfig struct {
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

type sendSumoLogic struct {
	conf sendSumoLogicConfig

	// client is safe for concurrent use.
	client http.HTTP
	// buffer is safe for concurrent use.
	mu     sync.Mutex
	buffer map[string]*aggregate.Bytes
}

func newSendSumoLogic(_ context.Context, cfg config.Config) (*sendSumoLogic, error) {
	conf := sendSumoLogicConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.URL == "" {
		return nil, fmt.Errorf("transform: send_sumologic: URL: %v", errors.ErrMissingRequiredOption)
	}

	send := sendSumoLogic{
		conf: conf,
	}

	send.client.Setup()
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		send.client.EnableXRay()
	}

	send.mu = sync.Mutex{}
	send.buffer = make(map[string]*aggregate.Bytes)

	return &send, nil
}

func (*sendSumoLogic) Close(context.Context) error {
	return nil
}

func (t *sendSumoLogic) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	t.mu.Lock()
	defer t.mu.Unlock()

	headers := []http.Header{
		{
			Key:   "Content-Type",
			Value: "application/json",
		},
	}

	var category string
	if t.conf.Category != "" {
		category = t.conf.Category
	}

	control := false
	for _, message := range messages {
		if message.IsControl() {
			control = true
			continue
		}

		if !json.Valid(message.Data()) {
			return nil, fmt.Errorf("transform: send_sumologic category %s: %v", category, errSendSumoLogicNonObject)
		}

		if t.conf.CategoryKey != "" {
			category = message.Get(t.conf.CategoryKey).String()
		}

		if _, ok := t.buffer[category]; !ok {
			// Aggregate up to 0.9MB or 10,000 items.
			// https://helt.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source#Data_payload_considerations
			t.buffer[category] = &aggregate.Bytes{}
			t.buffer[category].New(10000, 1000*1000*.9)
		}

		// Add data to the buffer. If buffer is full, then send the aggregated data.
		ok := t.buffer[category].Add(message.Data())
		if !ok {
			h := headers
			h = append(h, http.Header{
				Key:   "X-Sumo-Category",
				Value: category,
			})

			var buf bytes.Buffer
			items := t.buffer[category].Get()
			for _, i := range items {
				buf.WriteString(fmt.Sprintf("%s\n", i))
			}

			resp, err := t.client.Post(ctx, t.conf.URL, buf.Bytes(), h...)
			if err != nil {
				// Post errors returns metadata.
				return nil, fmt.Errorf("transform: send_sumologic: %v", err)
			}

			//nolint:errcheck // Response body is discarded to avoid resource leaks.
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			t.buffer[category].Reset()
			_ = t.buffer[category].Add(message.Data())
		}
	}

	// If a control message was received, then items are flushed from the buffer.
	if !control {
		return messages, nil
	}

	for category := range t.buffer {
		count := t.buffer[category].Count()
		if count == 0 {
			continue
		}

		h := headers
		h = append(h, http.Header{
			Key:   "X-Sumo-Category",
			Value: category,
		})

		var buf bytes.Buffer
		items := t.buffer[category].Get()
		for _, i := range items {
			buf.WriteString(fmt.Sprintf("%s\n", i))
		}

		resp, err := t.client.Post(ctx, t.conf.URL, buf.Bytes(), h...)
		if err != nil {
			// Post errors return metadata.
			return nil, fmt.Errorf("transform: send_sumologic: %v", err)
		}

		//nolint:errcheck // Response body is discarded to avoid resource leaks.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		delete(t.buffer, category)
	}

	return messages, nil
}
