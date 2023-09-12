package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNetworkIPPublic(_ context.Context, cfg config.Config) (*networkIPPublic, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPPublic{
		conf: conf,
	}

	return &insp, nil
}

type networkIPPublic struct {
	conf networkIPConfig
}

func (insp *networkIPPublic) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return netIPIsPublic(ip), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	ip := net.ParseIP(value.String())

	return netIPIsPublic(ip), nil
}

func (insp *networkIPPublic) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
