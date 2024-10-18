package condition

import (
	"context"
	"encoding/json"
	"net"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newNetworkIPUniqueLocalAddress(_ context.Context, cfg config.Config) (*networkIPUniqueLocalAddress, error) {
	conf := networkIPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := networkIPUniqueLocalAddress{
		conf: conf,
	}

	return &insp, nil
}

type networkIPUniqueLocalAddress struct {
	conf networkIPConfig
}

// Condition checks if the IP address is a unique local address (fd00::/8).
func (insp *networkIPUniqueLocalAddress) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		str := string(msg.Data())
		ip := net.ParseIP(str)

		return ip[0] == 0xfd, nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	ip := net.ParseIP(value.String())

	return ip[0] == 0xfd, nil
}

func (insp *networkIPUniqueLocalAddress) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
