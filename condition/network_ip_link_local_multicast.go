package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkIPLinkLocalMulticast(_ context.Context, cfg config.Config) (*networkIPLinkLocalMulticast, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPLinkLocalMulticast{
		conf: conf,
	}

	return &insp, nil
}

type networkIPLinkLocalMulticast struct {
	conf networkIPConfig
}

func (insp *networkIPLinkLocalMulticast) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip.IsLinkLocalMulticast(), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	ip := net.ParseIP(value.String())

	return ip.IsLinkLocalMulticast(), nil
}

func (insp *networkIPLinkLocalMulticast) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
