//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type enrichDNSIPLookup struct{}

func newEnrichDNSIPLookup(context.Context, config.Config) (*enrichDNS, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

func (*enrichDNSIPLookup) String() string {
	return ""
}

func (*enrichDNSIPLookup) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

type enrichDNSDomainLookup struct{}

func newEnrichDNSDomainLookup(context.Context, config.Config) (*enrichDNS, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

func (*enrichDNSDomainLookup) String() string {
	return ""
}

func (*enrichDNSDomainLookup) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

type enrichDNSTxtLookup struct{}

func newEnrichDNSTxtLookup(context.Context, config.Config) (*enrichDNS, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}

func (*enrichDNSTxtLookup) String() string {
	return ""
}

func (*enrichDNSTxtLookup) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_dns: %v", syscall.ENOSYS)
}
