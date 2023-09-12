package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkIPUnspecified(_ context.Context, cfg config.Config) (*networkIPUnspecified, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPUnspecified{
		conf: conf,
	}

	return &insp, nil
}

type networkIPUnspecified struct {
	conf networkIPConfig
}

func (insp *networkIPUnspecified) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return netIPIsUnspecified(ip), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	ip := net.ParseIP(value.String())

	return netIPIsUnspecified(ip), nil
}

func (c *networkIPUnspecified) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}