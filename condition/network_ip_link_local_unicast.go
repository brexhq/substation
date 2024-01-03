package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkIPLinkLocalUnicast(_ context.Context, cfg config.Config) (*networkIPLinkLocalUnicast, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPLinkLocalUnicast{
		conf: conf,
	}

	return &insp, nil
}

type networkIPLinkLocalUnicast struct {
	conf networkIPConfig
}

func (insp *networkIPLinkLocalUnicast) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip.IsLinkLocalUnicast(), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip.IsLinkLocalUnicast(), nil
}

func (insp *networkIPLinkLocalUnicast) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
