package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkIPUnicast(_ context.Context, cfg config.Config) (*networkIPUnicast, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPUnicast{
		conf: conf,
	}

	return &insp, nil
}

type networkIPUnicast struct {
	conf networkIPConfig
}

func (insp *networkIPUnicast) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip.IsGlobalUnicast() || ip.IsLinkLocalUnicast(), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip.IsGlobalUnicast() || ip.IsLinkLocalUnicast(), nil
}

func (insp *networkIPUnicast) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
