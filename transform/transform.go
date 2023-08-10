package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// errInvalidDataPattern is returned when a transform is configured with an invalid data access pattern. This is commonly caused by improperly set input and output settings.
var errInvalidDataPattern = fmt.Errorf("invalid data access pattern")

type Transformer interface {
	Transform(context.Context, ...*mess.Message) ([]*mess.Message, error)
	Close(context.Context) error
}

// NewTransformer returns a configured Transformer from a transform configuration.
func New(ctx context.Context, cfg config.Config) (Transformer, error) { //nolint: cyclop, gocyclo // ignore cyclomatic complexity
	switch cfg.Type {
	case "meta_for_each":
		return newMetaForEach(ctx, cfg)
	case "meta_pipeline":
		return newMetaPipeline(ctx, cfg)
	case "meta_plugin":
		return newMetaPlugin(ctx, cfg)
	case "meta_switch":
		return newMetaSwitch(ctx, cfg)
	case "proc_aws_dynamodb":
		return newProcAWSDynamoDB(ctx, cfg)
	case "proc_aws_lambda":
		return newProcAWSLambda(ctx, cfg)
	case "proc_base64":
		return newProcBase64(ctx, cfg)
	case "proc_capture":
		return newProcCapture(ctx, cfg)
	case "proc_case":
		return newProcCase(ctx, cfg)
	case "proc_combine":
		return newProcCombine(ctx, cfg)
	case "proc_convert":
		return newProcConvert(ctx, cfg)
	case "proc_copy":
		return newProcCopy(ctx, cfg)
	case "proc_delete":
		return newProcDelete(ctx, cfg)
	case "proc_dns":
		return newProcDNS(ctx, cfg)
	case "proc_domain":
		return newProcDomain(ctx, cfg)
	case "proc_drop":
		return newProcDrop(ctx, cfg)
	case "proc_err":
		return newProcErr(ctx, cfg)
	case "proc_expand":
		return newProcExpand(ctx, cfg)
	case "proc_flatten_array":
		return newProcFlattenArray(ctx, cfg)
	case "proc_group":
		return newProcGroup(ctx, cfg)
	case "proc_gzip":
		return newProcGzip(ctx, cfg)
	case "proc_hash":
		return newProcHash(ctx, cfg)
	case "proc_http":
		return newProcHTTP(ctx, cfg)
	case "proc_insert":
		return newProcInsert(ctx, cfg)
	case "proc_join":
		return newProcJoin(ctx, cfg)
	case "proc_jq":
		return newProcJQ(ctx, cfg)
	case "proc_kv_store":
		return newProcKVStore(ctx, cfg)
	case "proc_math":
		return newProcMath(ctx, cfg)
	case "proc_pretty_print":
		return newProcPrettyPrint(ctx, cfg)
	case "proc_replace":
		return newProcReplace(ctx, cfg)
	case "proc_split":
		return newProcSplit(ctx, cfg)
	case "proc_time":
		return newProcTime(ctx, cfg)
	case "send_aws_dynamodb":
		return newSendAWSDynamoDB(ctx, cfg)
	case "send_aws_kinesis":
		return newSendAWSKinesis(ctx, cfg)
	case "send_aws_kinesis_firehose":
		return newSendAWSKinesisFirehose(ctx, cfg)
	case "send_aws_s3":
		return newSendAWSS3(ctx, cfg)
	case "send_aws_sns":
		return newSendAWSSNS(ctx, cfg)
	case "send_aws_sqs":
		return newSendAWSSQS(ctx, cfg)
	case "send_file":
		return newSendFile(ctx, cfg)
	case "send_stdout":
		return newSendStdout(ctx, cfg)
	case "send_http":
		return newSendHTTP(ctx, cfg)
	case "send_sumologic":
		return newSendSumoLogic(ctx, cfg)
	default:
		return nil, fmt.Errorf("transform: new: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

func Apply(ctx context.Context, tforms []Transformer, messages ...*mess.Message) ([]*mess.Message, error) {
	resultMsgs := make([]*mess.Message, len(messages))
	copy(resultMsgs, messages)

	for i := 0; len(resultMsgs) > 0 && i < len(tforms); i++ {
		var nextResultMsgs []*mess.Message
		for _, m := range resultMsgs {
			rMsgs, err := tforms[i].Transform(ctx, m)
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
