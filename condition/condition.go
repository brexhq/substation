package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type Inspector interface {
	Inspect(context.Context, *message.Message) (bool, error)
}

func New(ctx context.Context, cfg config.Config) (Inspector, error) { //nolint: cyclop, gocyclo // ignore cyclomatic complexity
	switch cfg.Type {
	// Meta inspectors.
	case "all", "meta_all":
		return newMetaAll(ctx, cfg)
	case "any", "meta_any":
		return newMetaAny(ctx, cfg)
	case "none", "meta_none":
		return newMetaNone(ctx, cfg)
	// Format inspectors.
	case "format_mime":
		return newFormatMIME(ctx, cfg)
	case "format_json":
		return newFormatJSON(ctx, cfg)
	// Network inspectors.
	case "network_ip_global_unicast":
		return newNetworkIPGlobalUnicast(ctx, cfg)
	case "network_ip_link_local_multicast":
		return newNetworkIPLinkLocalMulticast(ctx, cfg)
	case "network_ip_link_local_unicast":
		return newNetworkIPLinkLocalUnicast(ctx, cfg)
	case "network_ip_loopback":
		return newNetworkIPLoopback(ctx, cfg)
	case "network_ip_multicast":
		return newNetworkIPMulticast(ctx, cfg)
	case "network_ip_private":
		return newNetworkIPPrivate(ctx, cfg)
	case "network_ip_unicast":
		return newNetworkIPUnicast(ctx, cfg)
	case "network_ip_unspecified":
		return newNetworkIPUnspecified(ctx, cfg)
	case "network_ip_valid":
		return newNetworkIPValid(ctx, cfg)
	// Number inspectors.
	case "number_equal_to":
		return newNumberEqualTo(ctx, cfg)
	case "number_less_than":
		return newNumberLessThan(ctx, cfg)
	case "number_greater_than":
		return newNumberGreaterThan(ctx, cfg)
	case "number_bitwise_and":
		return newNumberBitwiseAND(ctx, cfg)
	case "number_bitwise_or":
		return newNumberBitwiseOR(ctx, cfg)
	case "number_bitwise_xor":
		return newNumberBitwiseXOR(ctx, cfg)
	case "number_bitwise_not":
		return newNumberBitwiseNOT(ctx, cfg)
	case "number_length_less_than":
		return newNumberLengthLessThan(ctx, cfg)
	case "number_length_greater_than":
		return newNumberLengthGreaterThan(ctx, cfg)
	case "number_length_equal_to":
		return newNumberLengthEqualTo(ctx, cfg)
	// String inspectors.
	case "string_contains":
		return newStringContains(ctx, cfg)
	case "string_ends_with":
		return newStringEndsWith(ctx, cfg)
	case "string_equal_to":
		return newStringEqualTo(ctx, cfg)
	case "string_greater_than":
		return newStringGreaterThan(ctx, cfg)
	case "string_less_than":
		return newStringLessThan(ctx, cfg)
	case "string_starts_with":
		return newStringStartsWith(ctx, cfg)
	case "string_match":
		return newStringMatch(ctx, cfg)
	// Utility inspectors.
	case "utility_random":
		return newUtilityRandom(ctx, cfg)
	default:
		return nil, fmt.Errorf("condition %s: %w", cfg.Type, iconfig.ErrInvalidFactoryInput)
	}
}
