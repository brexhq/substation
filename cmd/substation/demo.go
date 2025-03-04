package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"
)

func init() {
	rootCmd.AddCommand(demoCmd)
}

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "demo substation",
	Long: `'substation demo' shows how Substation transforms data.
It prints an anonymized CloudTrail event (input) and the
transformed result (output) to the console. The event is 
partially normalized to the Elastic Common Schema (ECS).
`,
	// Examples:
	//  substation demo
	Example: `  substation demo
`,
	Args: cobra.MaximumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := compileStr(confDemo, nil)
		if err != nil {
			return err
		}

		cfg := substation.Config{}
		if err := json.Unmarshal([]byte(conf), &cfg); err != nil {
			return err
		}

		ctx := context.Background() // This doesn't need to be canceled.
		sub, err := substation.New(ctx, cfg)
		if err != nil {
			return err
		}

		msgs := []*message.Message{
			message.New().SetData([]byte(evtDemo)),
			message.New().AsControl(),
		}

		// Make the input pretty before printing to the console.
		fmt.Printf("input:\n%s\n", gjson.Get(evtDemo, "@this|@pretty").String())
		fmt.Printf("output:\n")

		if _, err := sub.Transform(ctx, msgs...); err != nil {
			return err
		}

		fmt.Printf("\nconfig:\n%s\n", confDemo)

		return nil
	},
}
