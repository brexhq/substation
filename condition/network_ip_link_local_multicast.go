package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
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

func (insp *networkIPLinkLocalMulticast) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.HasFlag(message.IsControl) {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip.IsLinkLocalMulticast(), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip.IsLinkLocalMulticast(), nil
}

func (insp *networkIPLinkLocalMulticast) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
