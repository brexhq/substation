package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newNetworkIPValid(_ context.Context, cfg config.Config) (*networkIPValid, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPValid{
		conf: conf,
	}

	return &insp, nil
}

type networkIPValid struct {
	conf networkIPConfig
}

func (insp *networkIPValid) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.HasFlag(message.IsControl) {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip != nil, nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip != nil, nil
}

func (insp *networkIPValid) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
