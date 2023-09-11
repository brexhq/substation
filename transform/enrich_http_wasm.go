//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newEnrichHTTPGet(context.Context, config.Config) (*enrichHTTPGet, error) {
	return nil, fmt.Errorf("transform: enrich_http: %v", syscall.ENOSYS)
}

type enrichHTTPGet struct{}

func (*enrichHTTPGet) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_http: %v", syscall.ENOSYS)
}

func (*enrichHTTPGet) String() string {
	return ""
}

func newEnrichHTTPPost(context.Context, config.Config) (*enrichHTTPPost, error) {
	return nil, fmt.Errorf("transform: enrich_http: %v", syscall.ENOSYS)
}

type enrichHTTPPost struct{}

func (*enrichHTTPPost) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_http: %v", syscall.ENOSYS)
}

func (*enrichHTTPPost) String() string {
	return ""
}
