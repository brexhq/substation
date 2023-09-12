package condition

import (
	"context"
	gojson "encoding/json"
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

	if insp.conf.Object.Key == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return netIPIsUnicast(ip), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	ip := net.ParseIP(value.String())

	return netIPIsUnicast(ip), nil
}

func (insp *networkIPUnicast) String() string {
	b, _ := gojson.Marshal(insp.conf)
	return string(b)
}
