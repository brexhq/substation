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

	transformer Transformer
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

	tform, ok := sym.(Transformer)
	if !ok {
		return nil, fmt.Errorf("transform: meta_plugin: %v", errMetaPluginInterfaceNotImplemented)
	}

	if err := _config.Decode(conf.Settings, tform); err != nil {
		return nil, err
	}

	meta := metaPlugin{
		conf:        conf,
		transformer: tform,
	}

	return &meta, nil
}

func (t *metaPlugin) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (t *metaPlugin) Close(ctx context.Context) error {
	return t.transformer.Close(ctx)
}

func (t *metaPlugin) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	msgs, err := t.transformer.Transform(ctx, messages...)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_plugin: %v", err)
	}

	return msgs, nil
}
