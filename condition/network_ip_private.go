package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkIPPrivate(_ context.Context, cfg config.Config) (*networkIPPrivate, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPPrivate{
		conf: conf,
	}

	return &insp, nil
}

type networkIPPrivate struct {
	conf networkIPConfig
}

func (insp *networkIPPrivate) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return netIPIsPrivate(ip), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	ip := net.ParseIP(value.String())

	return netIPIsPrivate(ip), nil
}

func (insp *networkIPPrivate) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
