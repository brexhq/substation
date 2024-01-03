package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkIPGlobalUnicast(_ context.Context, cfg config.Config) (*networkIPGlobalUnicast, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPGlobalUnicast{
		conf: conf,
	}

	return &insp, nil
}

type networkIPGlobalUnicast struct {
	conf networkIPConfig
}

func (insp *networkIPGlobalUnicast) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip.IsGlobalUnicast(), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip.IsGlobalUnicast(), nil
}

func (insp *networkIPGlobalUnicast) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
