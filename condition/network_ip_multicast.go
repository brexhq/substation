package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
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

func (insp *networkIPMulticast) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return netIPIsMulticast(ip), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	ip := net.ParseIP(value.String())

	return netIPIsMulticast(ip), nil
}

func (insp *networkIPMulticast) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
