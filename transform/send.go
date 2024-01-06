package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/message"
)

// errSendBatchMisconfigured is returned when data cannot be successfully added
// to a batch. This is usually due to a misconfiguration, such as a size, count,
// or duration limit.
var errSendBatchMisconfigured = fmt.Errorf("data could not be added to batch")

func withTransforms(ctx context.Context, tf []Transformer, items [][]byte) ([][]byte, error) {
	if tf == nil {
		return items, nil
	}

	var msg []*message.Message
	for _, i := range items {
		msg = append(msg, message.New().SetData(i))
	}
	msg = append(msg, message.New(message.AsControl()))

	res, err := Apply(ctx, tf, msg...)
	if err != nil {
		return nil, err
	}

	var output [][]byte
	for _, r := range res {
		if r.IsControl() {
			continue
		}

		output = append(output, r.Data())
	}

	return output, nil
}
