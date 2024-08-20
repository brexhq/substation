package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
	"github.com/tidwall/gjson"
)

func newAggregateFromArray(_ context.Context, cfg config.Config) (*aggregateFromArray, error) {
	conf := aggregateArrayConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform aggregate_from_array: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "aggregate_from_array"
	}

	tf := aggregateFromArray{
		conf:      conf,
		hasObjSrc: conf.Object.SourceKey != "",
		hasObjTrg: conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type aggregateFromArray struct {
	conf      aggregateArrayConfig
	hasObjSrc bool
	hasObjTrg bool
}

func (tf *aggregateFromArray) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	meta := msg.Metadata()
	var output []*message.Message

	var value message.Value
	if tf.hasObjSrc {
		value = msg.GetValue(tf.conf.Object.SourceKey)
		if err := msg.DeleteValue(tf.conf.Object.SourceKey); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	for _, res := range value.Array() {
		outMsg := message.New().SetMetadata(meta)

		if tf.hasObjSrc {
			for key, val := range msg.GetValue("@this").Map() {
				if err := outMsg.SetValue(key, val.Value()); err != nil {
					return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
				}
			}
		}

		if tf.hasObjTrg {
			if err := outMsg.SetValue(tf.conf.Object.TargetKey, res); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}

			output = append(output, outMsg)
			continue
		}

		if tf.hasObjSrc {
			tmp := fmt.Sprintf(`[%s,%s]`, outMsg.Data(), res.String())
			join := gjson.GetBytes([]byte(tmp), "@join")

			outMsg.SetData([]byte(join.String()))
			output = append(output, outMsg)
			continue
		}

		outMsg.SetData(res.Bytes())
		output = append(output, outMsg)
	}

	return output, nil
}

func (tf *aggregateFromArray) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
