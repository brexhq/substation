package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
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

func (insp *networkIPUnspecified) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip.IsUnspecified(), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip.IsUnspecified(), nil
}

func (c *networkIPUnspecified) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
