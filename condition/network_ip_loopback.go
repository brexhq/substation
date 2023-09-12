package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkIPLoopback(_ context.Context, cfg config.Config) (*networkIPLoopback, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPLoopback{
		conf: conf,
	}

	return &insp, nil
}

type networkIPLoopback struct {
	conf networkIPConfig
}

func (insp *networkIPLoopback) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return netIPIsLoopback(ip), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	ip := net.ParseIP(value.String())

	return netIPIsLoopback(ip), nil
}

func (insp *networkIPLoopback) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}