//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type enrichDNSFwdLookup struct{}

func newEnrichDNSFwdLookup(context.Context, config.Config) (*enrichDNS, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

func (*enrichDNSFwdLookup) String() string {
	return ""
}

func (*enrichDNSFwdLookup) Close(context.Context) error {
	return nil
}

func (*enrichDNSFwdLookup) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

type enrichDNSRevLookup struct{}

func newEnrichDNSRevLookup(context.Context, config.Config) (*enrichDNS, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

func (*enrichDNSRevLookup) String() string {
	return ""
}

func (*enrichDNSRevLookup) Close(context.Context) error {
	return nil
}

func (*enrichDNSRevLookup) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

type enrichDNSTxtLookup struct{}

func newEnrichDNSTxtLookup(context.Context, config.Config) (*enrichDNS, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

func (*enrichDNSTxtLookup) String() string {
	return ""
}

func (*enrichDNSTxtLookup) Close(context.Context) error {
	return nil
}

func (*enrichDNSTxtLookup) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}
