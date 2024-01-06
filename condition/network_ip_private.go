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

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip.IsPrivate(), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip.IsPrivate(), nil
}

func (insp *networkIPPrivate) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
