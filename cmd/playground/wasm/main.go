//go:build wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
	"github.com/google/go-jsonnet"
)

var vm = jsonnet.MakeVM()

func main() {
	js.Global().Get("window").Set("play", js.FuncOf(
		func(this js.Value, args []js.Value) interface{} {
			sonnet := args[0].String()
			events := args[1].String()
			library := args[2].String()

			// Concatenate the library with the user's configuration.
			sonnet = fmt.Sprintf("local sub = %s; \n\n%s", library, sonnet)
			cfg, err := vm.EvaluateAnonymousSnippet("", sonnet)
			if err != nil {
				return js.ValueOf(fmt.Sprintf("jsonnet: %s", err.Error()))
			}

			// Configuration must be an array.
			conf := []config.Config{}
			if err := json.Unmarshal([]byte(cfg), &conf); err != nil {
				return js.ValueOf(fmt.Sprintf("unmarshal: %s", err.Error()))
			}

			tforms, err := transform.NewTransformers(context.Background(), conf...)
			if err != nil {
				return js.ValueOf(fmt.Sprintf("substation: %s", err.Error()))
			}

			// each line of data is treated as a separate input into the processors
			var msgs []*message.Message
			events = strings.TrimSpace(events)
			for _, e := range strings.Split(events, "\n") {
				if e == "" {
					continue
				}

				data, err := message.New(
					message.SetData([]byte(e)),
				)
				if err != nil {
					return js.ValueOf(fmt.Sprintf("message: %s", err.Error()))
				}

				msgs = append(msgs, data)
			}

			msgs, err = transform.Apply(context.Background(), tforms, msgs...)
			if err != nil {
				return js.ValueOf(fmt.Sprintf("transform: %s", err.Error()))
			}

			// each processed output is returned on a new line
			var s string
			for i, msg := range msgs {
				s += string(msg.Data())
				if i < len(msgs)-1 {
					s += "\n"
				}
			}

			return js.ValueOf(s)
		},
	))

	select {}
}
