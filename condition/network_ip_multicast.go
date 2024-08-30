package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newNetworkIPMulticast(_ context.Context, cfg config.Config) (*networkIPMulticast, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPMulticast{
		conf: conf,
	}

	return &insp, nil
}

type networkIPMulticast struct {
	conf networkIPConfig
}

func (insp *networkIPMulticast) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip.IsMulticast(), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip.IsMulticast(), nil
}

func (insp *networkIPMulticast) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
