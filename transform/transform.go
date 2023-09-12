// package transform provides capabilities for transforming data.
package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// Transformer is the interface implemented by all transforms and
// provides the ability to transform data.
type Transformer interface {
	Transform(context.Context, *message.Message) ([]*message.Message, error)
}

// New returns a configured Transformer.
func New(ctx context.Context, cfg config.Config) (Transformer, error) { //nolint: cyclop, gocyclo // ignore cyclomatic complexity
	switch cfg.Type {
	// Meta transforms.
	case "meta_err":
		return newMetaErr(ctx, cfg)
	case "meta_for_each":
		return newMetaForEach(ctx, cfg)
	case "meta_pipeline":
		return newMetaPipeline(ctx, cfg)
	case "meta_plugin":
		return newMetaPlugin(ctx, cfg)
	case "meta_switch":
		return newMetaSwitch(ctx, cfg)
	// Aggregation transforms.
	case "aggregate_from_array":
		return newAggregateFromArray(ctx, cfg)
	case "aggregate_to_array":
		return newAggregateToArray(ctx, cfg)
	case "aggregate_from_str":
		return newAggregateFromStr(ctx, cfg)
	case "aggregate_to_str":
		return newAggregateToStr(ctx, cfg)
	// Array transforms.
	case "array_group":
		return newArrayGroup(ctx, cfg)
	case "array_join":
		return newArrayJoin(ctx, cfg)
	// Compress transforms.
	case "compress_from_gzip":
		return newCompressFromGzip(ctx, cfg)
	case "compress_to_gzip":
		return newCompressToGzip(ctx, cfg)
	// Enrichment transforms.
	case "enrich_aws_dynamodb":
		return newEnrichAWSDynamoDB(ctx, cfg)
	case "enrich_aws_lambda":
		return newEnrichAWSLambda(ctx, cfg)
	case "enrich_dns_forward_lookup":
		return newEnrichDNSIPLookup(ctx, cfg)
	case "enrich_dns_reverse_lookup":
		return newEnrichDNSDomainLookup(ctx, cfg)
	case "enrich_dns_text_lookup":
		return newEnrichDNSTxtLookup(ctx, cfg)
	case "enrich_http_get":
		return newEnrichHTTPGet(ctx, cfg)
	case "enrich_http_post":
		return newEnrichHTTPPost(ctx, cfg)
	case "enrich_kv_store_get":
		return newEnrichKVStoreGet(ctx, cfg)
	case "enrich_kv_store_set":
		return newEnrichKVStoreSet(ctx, cfg)
	// External transforms.
	case "external_jq":
		return newExternalJQ(ctx, cfg)
	// Format transforms.
	case "format_from_base64":
		return newFormatFromBase64(ctx, cfg)
	case "format_to_base64":
		return newFormatToBase64(ctx, cfg)
	case "format_from_pretty_print":
		return newFormatFromPrettyPrint(ctx, cfg)
	// Hash transforms.
	case "hash_md5":
		return newHashMD5(ctx, cfg)
	case "hash_sha256":
		return newHashSHA256(ctx, cfg)
	// Logic transforms.
	case "logic_num_add":
		return newLogicNumAdd(ctx, cfg)
	case "logic_num_divide":
		return newLogicNumDivide(ctx, cfg)
	case "logic_num_multiply":
		return newLogicNumMultiply(ctx, cfg)
	case "logic_num_subtract":
		return newLogicNumSubtract(ctx, cfg)
	// Network transforms.
	case "network_domain_registered_domain":
		return newNetworkDomainRegisteredDomain(ctx, cfg)
	case "network_domain_subdomain":
		return newNetworkDomainSubdomain(ctx, cfg)
	case "network_domain_top_level_domain":
		return newNetworkDomainTopLevelDomain(ctx, cfg)
	// Object transforms.
	case "object_copy":
		return newObjectCopy(ctx, cfg)
	case "object_delete":
		return newObjectDelete(ctx, cfg)
	case "object_insert":
		return newObjectInsert(ctx, cfg)
	case "object_to_bool":
		return newObjectToBool(ctx, cfg)
	case "object_to_float":
		return newObjectToFloat(ctx, cfg)
	case "object_to_int":
		return newObjectToInt(ctx, cfg)
	case "object_to_str":
		return newObjectToStr(ctx, cfg)
	case "object_to_uint":
		return newObjectToUint(ctx, cfg)
	// Send transforms.
	case "send_aws_dynamodb":
		return newSendAWSDynamoDB(ctx, cfg)
	case "send_aws_kinesis_firehose":
		return newSendAWSKinesisDataFirehose(ctx, cfg)
	case "send_aws_kinesis":
		return newSendAWSKinesisDataStream(ctx, cfg)
	case "send_aws_s3":
		return newSendAWSS3(ctx, cfg)
	case "send_aws_sns":
		return newSendAWSSNS(ctx, cfg)
	case "send_aws_sqs":
		return newSendAWSSQS(ctx, cfg)
	case "send_file":
		return newSendFile(ctx, cfg)
	case "send_http":
		return newSendHTTP(ctx, cfg)
	case "send_stdout":
		return newSendStdout(ctx, cfg)
	case "send_sumologic":
		return newSendSumologic(ctx, cfg)
	// String transforms.
	case "string_pattern_find_all":
		return newStringPatternFindAll(ctx, cfg)
	case "string_pattern_find":
		return newStringPatternFind(ctx, cfg)
	case "string_pattern_named_group":
		return newStringPatternNamedGroup(ctx, cfg)
	case "string_to_lower":
		return newStringToLower(ctx, cfg)
	case "string_to_snake":
		return newStringToSnake(ctx, cfg)
	case "string_to_upper":
		return newStringToUpper(ctx, cfg)
	case "string_replace":
		return newStringReplace(ctx, cfg)
	case "string_split":
		return newStringSplit(ctx, cfg)
	// Time transforms.
	case "time_from_str":
		return newTimeFromStr(ctx, cfg)
	case "time_from_unix":
		return newTimeFromUnix(ctx, cfg)
	case "time_now":
		return newTimeNow(ctx, cfg)
	case "time_to_str":
		return newTimeToStr(ctx, cfg)
	case "time_to_unix":
		return newTimeToUnix(ctx, cfg)
	// Utility transforms.
	case "utility_delay":
		return newUtilityDelay(ctx, cfg)
	case "utility_drop":
		return newUtilityDrop(ctx, cfg)
	case "utility_err":
		return newUtilityErr(ctx, cfg)
	default:
		return nil, fmt.Errorf("transform: new: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

func Apply(ctx context.Context, tf []Transformer, mess ...*message.Message) ([]*message.Message, error) {
	resultMsgs := make([]*message.Message, len(mess))
	copy(resultMsgs, mess)

	for i := 0; len(resultMsgs) > 0 && i < len(tf); i++ {
		var nextResultMsgs []*message.Message
		for _, m := range resultMsgs {
			rMsgs, err := tf[i].Transform(ctx, m)
			if err != nil {
				// We immediately return if a transform hits an unrecoverable
				// error on a message.
				return nil, err
			}
			nextResultMsgs = append(nextResultMsgs, rMsgs...)
		}
		resultMsgs = nextResultMsgs
	}

	return resultMsgs, nil
}

func bytesToValue(b []byte) message.Value {
	msg := message.New()
	msg.SetValue("_", b)

	return msg.GetValue("_")
}
