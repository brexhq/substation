package main

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/message"
)

// Duplicates a message.
type Duplicate struct {
	// Count is the number of times to duplicate the message.
	Count int `json:"count"`
}

// Transforms a message based on the configuration.
func (t *Duplicate) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Always return control messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	output := []*message.Message{msg}
	for i := 0; i < t.Count; i++ {
		output = append(output, msg)
	}

	return output, nil
}

func main() {
	// Create a data message.
	b := []byte(`{"a":"b"}`)
	msg := message.New().SetData(b)

	// Create the transform.
	tf := Duplicate{Count: 2}

	// Run the transform.
	ctx := context.Background()
	msgs, err := tf.Transform(ctx, msg)
	if err != nil {
		// Handle error.
		panic(err)
	}

	// Print the output.
	for _, m := range msgs {
		fmt.Println(string(m.Data()))
	}
}
