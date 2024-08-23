package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/secrets"
	"github.com/brexhq/substation/v2/message"
)

type utilitySecretConfig struct {
	// Secret is the secret to retrieve.
	Secret config.Config `json:"secret"`

	ID string `json:"id"`
}

func (c *utilitySecretConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilitySecret(ctx context.Context, cfg config.Config) (*utilitySecret, error) {
	// conf gets validated when calling secrets.New.
	conf := utilitySecretConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform utility_secret: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "utility_secret"
	}

	ret, err := secrets.New(ctx, conf.Secret)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := utilitySecret{
		conf:   conf,
		secret: ret,
	}

	if err := tf.secret.Retrieve(ctx); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	return &tf, nil
}

type utilitySecret struct {
	conf utilitySecretConfig

	// secret is safe for concurrent access.
	secret secrets.Retriever
}

func (tf *utilitySecret) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if tf.secret.Expired() {
		if err := tf.secret.Retrieve(ctx); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *utilitySecret) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
