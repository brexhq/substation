package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
	"github.com/tidwall/gjson"
)

func newAggregateFromArray(_ context.Context, cfg config.Config) (*aggregateFromArray, error) {
	conf := aggregateArrayConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_aggregate_from_array: %v", err)
	}

	tf := aggregateFromArray{
		conf:         conf,
		hasObjKey:    conf.Object.Key != "",
		hasObjSetKey: conf.Object.SetKey != "",
	}

	return &tf, nil
}

type aggregateFromArray struct {
	conf         aggregateArrayConfig
	hasObjKey    bool
	hasObjSetKey bool
}

func (tf *aggregateFromArray) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	meta := msg.Metadata()
	var output []*message.Message

	var value message.Value
	if tf.hasObjKey {
		value = msg.GetValue(tf.conf.Object.Key)
		if err := msg.DeleteValue(tf.conf.Object.Key); err != nil {
			return nil, err
		}
	} else {
		value = bytesToValue(msg.Data())
	}

	for _, res := range value.Array() {
		outMsg := message.New().SetMetadata(meta)

		if tf.hasObjKey {
			for key, val := range msg.GetValue("@this").Map() {
				if err := outMsg.SetValue(key, val.Value()); err != nil {
					return nil, err
				}
			}
		}

		if tf.hasObjSetKey {
			if err := outMsg.SetValue(tf.conf.Object.SetKey, res); err != nil {
				return nil, err
			}

			output = append(output, outMsg)
			continue
		}

		if tf.hasObjKey {
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

func (*aggregateFromArray) Close(context.Context) error {
	return nil
}
