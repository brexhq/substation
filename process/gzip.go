package process

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Gzip processes data by compressing or decompressing gzip. The processor supports these patterns:
	data:
		[31 139 8 0 0 0 0 0 0 255 74 203 207 7 4 0 0 255 255 33 101 115 140 3 0 0 0] >>> foo
		foo >>> [31 139 8 0 0 0 0 0 0 255 74 203 207 7 4 0 0 255 255 33 101 115 140 3 0 0 0]

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "gzip",
		"settings": {
			"options": {
				"direction": "from"
			}
		}
	}
*/
type Gzip struct {
	Options   GzipOptions      `json:"options"`
	Condition condition.Config `json:"condition"`
}

/*
GzipOptions contains custom options settings for the Gzip processor:
	Direction:
		the direction of the compression
		must be one of:
			to: compress data to gzip
			from: decompress data from gzip
*/
type GzipOptions struct {
	Direction string `json:"direction"`
}

func (p Gzip) from(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: %v", err)
	}

	output, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("gzip: %v", err)
	}

	return output, nil
}

func (p Gzip) to(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("gzip: %v", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzip: %v", err)
	}

	return buf.Bytes(), nil
}

// ApplyBatch processes a slice of encapsulated data with the Gzip processor. Conditions are optionally applied to the data to enable processing.
func (p Gzip) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %v", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %v", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Gzip processor.
func (p Gzip) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return cap, fmt.Errorf("apply settings %+v: %w", p, ProcessorInvalidSettings)
	}

	switch s := p.Options.Direction; s {
	case "from":
		tmp, err := p.from(cap.GetData())
		if err != nil {
			return cap, fmt.Errorf("apply settings %+v: %v", p, err)
		}

		cap.SetData(tmp)
		return cap, nil
	case "to":
		tmp, err := p.to(cap.GetData())
		if err != nil {
			return cap, fmt.Errorf("apply settings %+v: %v", p, err)
		}

		cap.SetData(tmp)
		return cap, nil
	}

	return cap, fmt.Errorf("apply settings %+v: %w", p, ProcessorInvalidSettings)
}
