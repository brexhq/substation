package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"plugin"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

var errMetaPluginInterfaceNotImplemented = fmt.Errorf("plugin does not implement transformer interface")

type metaPluginConfig struct {
	Plugin   string                 `json:"plugin"`
	Settings map[string]interface{} `json:"settings"`
}

type metaPlugin struct {
	conf metaPluginConfig

	tf Transformer
}

func newMetaPlugin(_ context.Context, cfg config.Config) (*metaPlugin, error) {
	conf := metaPluginConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: meta_plugin: %v", err)
	}

	plug, err := plugin.Open(conf.Plugin)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_plugin: %v", err)
	}

	sym, err := plug.Lookup("Transformer")
	if err != nil {
		return nil, fmt.Errorf("transform: meta_plugin: %v", err)
	}

	tf, ok := sym.(Transformer)
	if !ok {
		return nil, fmt.Errorf("transform: meta_plugin: %v", errMetaPluginInterfaceNotImplemented)
	}

	if err := _config.Decode(conf.Settings, tf); err != nil {
		return nil, err
	}

	meta := metaPlugin{
		conf: conf,
		tf:   tf,
	}

	return &meta, nil
}

func (meta *metaPlugin) String() string {
	b, _ := gojson.Marshal(meta.conf)
	return string(b)
}

func (meta *metaPlugin) Close(ctx context.Context) error {
	return meta.tf.Close(ctx)
}

func (meta *metaPlugin) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	msgs, err := meta.tf.Transform(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_plugin: %v", err)
	}

	return msgs, nil
}
