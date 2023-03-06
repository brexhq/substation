//go:build wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
	"github.com/google/go-jsonnet"
)

var vm = jsonnet.MakeVM()

func main() {
	js.Global().Get("window").Set("play", js.FuncOf(
		func(this js.Value, args []js.Value) interface{} {
			sonnet := args[0].String()
			events := args[1].String()
			library := args[2].String()

			// concatenate the library with the user's configuration
			sonnet = fmt.Sprintf("local sub = %s; \n\n%s", library, sonnet)
			cfg, err := vm.EvaluateAnonymousSnippet("", sonnet)
			if err != nil {
				return js.ValueOf(fmt.Sprintf("jsonnet: %s", err.Error()))
			}

			// configuration must be an array
			conf := []config.Config{}
			if err := json.Unmarshal([]byte(cfg), &conf); err != nil {
				return js.ValueOf(fmt.Sprintf("unmarshal: %s", err.Error()))
			}

			batchers, err := process.NewBatchers(conf...)
			if err != nil {
				return js.ValueOf(fmt.Sprintf("substation: %s", err.Error()))
			}

			// each line of data is treated as a separate input into the processors
			var b [][]byte
			events = strings.TrimSpace(events)
			for _, e := range strings.Split(events, "\n") {
				b = append(b, []byte(e))
			}

			batch, err := process.BatchBytes(context.TODO(), b, batchers...)
			if err != nil {
				return js.ValueOf(fmt.Sprintf("substation: %s", err.Error()))
			}

			// each processed output is returned on a new line
			var s string
			for i, b := range batch {
				s += string(b)
				if i < len(batch)-1 {
					s += "\n"
				}
			}

			return js.ValueOf(s)
		},
	))

	select {}
}
