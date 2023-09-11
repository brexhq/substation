package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"plugin"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

var errMetaPluginInterfaceNotImplemented = fmt.Errorf("plugin does not implement transformer interface")

type metaPluginConfig struct {
	Plugin   string                 `json:"plugin"`
	Settings map[string]interface{} `json:"settings"`
}

func (c *metaPluginConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaPluginConfig) Validate() error {
	if c.Plugin == "" {
		return fmt.Errorf("plugin: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type metaPlugin struct {
	conf metaPluginConfig

	tf Transformer
}

func newMetaPlugin(_ context.Context, cfg config.Config) (*metaPlugin, error) {
	conf := metaPluginConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_meta_plugin: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_meta_plugin: %v", err)
	}

	plug, err := plugin.Open(conf.Plugin)
	if err != nil {
		return nil, fmt.Errorf("transform: new_meta_plugin: %v", err)
	}

	sym, err := plug.Lookup("Transformer")
	if err != nil {
		return nil, fmt.Errorf("transform: new_meta_plugin: %v", err)
	}

	tf, ok := sym.(Transformer)
	if !ok {
		return nil, fmt.Errorf("transform: new_meta_plugin: %v", errMetaPluginInterfaceNotImplemented)
	}

	if err := iconfig.Decode(conf.Settings, tf); err != nil {
		return nil, err
	}

	meta := metaPlugin{
		conf: conf,
		tf:   tf,
	}

	return &meta, nil
}

func (meta *metaPlugin) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	msgs, err := meta.tf.Transform(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_plugin: %v", err)
	}

	return msgs, nil
}

func (meta *metaPlugin) String() string {
	b, _ := json.Marshal(meta.conf)
	return string(b)
}
