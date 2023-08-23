package transform

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
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
	Buffer aggregate.Config `json:"buffer"`
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
	client  http.HTTP
	headers []http.Header
	// buffer is safe for concurrent use.
	mu        sync.Mutex
	buffer    map[string]*aggregate.Aggregate
	bufferCfg aggregate.Config
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

	send.headers = []http.Header{
		{
			Key:   "Content-Type",
			Value: "application/json",
		},
	}

	send.mu = sync.Mutex{}
	send.buffer = make(map[string]*aggregate.Aggregate)
	send.bufferCfg = aggregate.Config{
		// SumoLogic limits batches to 1MB.
		Size:     1024 * 1024,
		Count:    conf.Buffer.Count,
		Interval: conf.Buffer.Interval,
	}

	return &send, nil
}

func (*sendSumoLogic) Close(context.Context) error {
	return nil
}

func (send *sendSumoLogic) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	send.mu.Lock()
	defer send.mu.Unlock()

	if message.IsControl() {
		for category := range send.buffer {
			count := send.buffer[category].Count()
			if count == 0 {
				continue
			}

			if err := send.sendPayload(ctx, category); err != nil {
				return nil, fmt.Errorf("transform: send_sumologic: %v", err)
			}
		}

		send.buffer = make(map[string]*aggregate.Aggregate)
		return []*mess.Message{message}, nil
	}

	var category string
	if send.conf.Category != "" {
		category = send.conf.Category
	}

	if !json.Valid(message.Data()) {
		return nil, fmt.Errorf("transform: send_sumologic category %s: %v", category, errSendSumoLogicNonObject)
	}

	if send.conf.CategoryKey != "" {
		category = message.Get(send.conf.CategoryKey).String()
	}

	if _, ok := send.buffer[category]; !ok {
		agg, err := aggregate.New(send.bufferCfg)
		if err != nil {
			return nil, fmt.Errorf("transform: send_sumologic: %v", err)
		}

		send.buffer[category] = agg
	}

	// Sends data to SumoLogic only when the buffer is full.
	if ok := send.buffer[category].Add(message.Data()); ok {
		return []*mess.Message{message}, nil
	}

	if err := send.sendPayload(ctx, category); err != nil {
		return nil, fmt.Errorf("transform: send_sumologic: %v", err)
	}

	// Reset the buffer and add the message data.
	send.buffer[category].Reset()
	_ = send.buffer[category].Add(message.Data())

	return []*mess.Message{message}, nil
}

func (t *sendSumoLogic) sendPayload(ctx context.Context, category string) error {
	if t.buffer[category].Count() == 0 {
		return nil
	}

	h := t.headers
	h = append(h, http.Header{
		Key:   "X-Sumo-Category",
		Value: category,
	})

	var buf bytes.Buffer
	for _, i := range t.buffer[category].Get() {
		buf.WriteString(fmt.Sprintf("%s\n", i))
	}

	resp, err := t.client.Post(ctx, t.conf.URL, buf.Bytes(), h...)
	if err != nil {
		// Post errors return metadata.
		return err
	}

	//nolint:errcheck // Response body is discarded to avoid resource leaks.
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return nil
}
