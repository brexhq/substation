package transform

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
	"github.com/google/uuid"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
)

// errInvalidDataPattern is returned when a transform is configured with an invalid data access pattern. This is commonly caused by improperly set input and output settings.
var errInvalidDataPattern = fmt.Errorf("invalid data access pattern")

type Transformer interface {
	Transform(context.Context, ...*mess.Message) ([]*mess.Message, error)
	Close(context.Context) error
}

// NewTransformer returns a configured Transformer from a transform configuration.
func NewTransformer(ctx context.Context, cfg config.Config) (Transformer, error) { //nolint: cyclop, gocyclo // ignore cyclomatic complexity
	switch cfg.Type {
	case "meta_for_each":
		return newMetaForEach(ctx, cfg)
	case "meta_pipeline":
		return newMetaPipeline(ctx, cfg)
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
	case "proc_condense":
		return newProcCondense(ctx, cfg)
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
	case "proc_error":
		return newProcError(ctx, cfg)
	case "proc_expand":
		return newProcExpand(ctx, cfg)
	case "proc_flatten":
		return newProcFlatten(ctx, cfg)
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
		return nil, fmt.Errorf("process: new_transformer: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

// NewTransformers accepts one or more transform configurations and returns configured batchers.
func NewTransformers(ctx context.Context, cfg ...config.Config) ([]Transformer, error) {
	var bats []Transformer

	for _, c := range cfg {
		b, err := NewTransformer(ctx, c)
		if err != nil {
			return nil, err
		}

		bats = append(bats, b)
	}

	return bats, nil
}

// CloseTransformers closes all batchers and returns an error if any close fails.
func CloseTransformers(ctx context.Context, batchers ...Transformer) error {
	for _, b := range batchers {
		if err := b.Close(ctx); err != nil {
			return err
		}
	}

	return nil
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

type fw struct {
	*os.File
	w       io.WriteCloser
	newline []byte
}

func (f *fw) Write(b []byte) (int, error) {
	if f.newline != nil {
		b = append(b, f.newline...)
	}

	if f.w != nil {
		return f.w.Write(b)
	}

	return f.File.Write(b)
}

func (f *fw) Close() error {
	if f.w != nil {
		if err := f.w.Close(); err != nil {
			return err
		}
	}

	return f.File.Close()
}

func NewFileWrapper(f *os.File, format config.Config, compression config.Config) (*fw, error) {
	var newline []byte
	switch runtime.GOOS {
	case "windows":
		newline = []byte("\r\n")
	default:
		newline = []byte("\n")
	}

	// if the file format is not text-based, then newline is unused.
	// if a file format uses a specific compression, then it should
	// be configured and returned in this switch.
	switch format.Type {
	case "data":
		newline = nil
	}

	switch compression.Type {
	case "gzip":
		return &fw{f, gzip.NewWriter(f), newline}, nil
	case "snappy":
		return &fw{f, snappy.NewBufferedWriter(f), newline}, nil
	case "zstd":
		// TODO: add settings support
		z, err := zstd.NewWriter(f)
		if err != nil {
			return nil, err
		}

		return &fw{f, z, newline}, nil
	default:
		return &fw{f, nil, newline}, nil
	}
}

type filePath struct {
	// Prefix is a prefix prepended to the file path.
	//
	// This is optional and has no default.
	Prefix string `json:"prefix"`
	// PrefixKey retrieves a value from an object that is used as
	// the prefix prepended to the file path. If used, then
	// this overrides Prefix.
	//
	// This is optional and has no default.
	PrefixKey string `json:"prefix_key"`
	// Suffix is a suffix appended to the file path and is used as
	// the object filename.
	//
	// This is optional and has no default.
	Suffix string `json:"suffix"`
	// SuffixKey retrieves a value from an object that is used as
	// the suffix appended to the file path. If used, then
	// this overrides Suffix.
	//
	// This is optional and has no default.
	SuffixKey string `json:"suffix_key"`
	// TimeFormat inserts a formatted datetime string into the file path.
	// Must be one of:
	//
	// - pattern-based layouts (https://gobyexample.com/procTime-formatting-parsing)
	//
	// - unix: epoch (supports fractions of a second)
	//
	// - unix_milli: epoch milliseconds
	//
	// This is optional and has no default.
	TimeFormat string `json:"time_format"`
	// UUID inserts a random UUID into the file path. If a suffix is
	// not set, then this is used as the filename.
	//
	// This is optional and defaults to false.
	UUID bool `json:"uuid"`
	// Extension appends a file extension to the filename.
	//
	// This is optional and defaults to false.
	Extension bool `json:"extension"`
}

// New constructs a file path using the pattern
// [prefix]/[time_format]/[uuid]/[suffix], where each field is optional
// and builds on the previous field. The caller is responsible for
// creating an OS agnostic file path (filepath.FromSlash is recommended).
//
// If only one field is set, then this constructs a filename,
// otherwise it constructs a file path.
//
// If the struct is empty, then this returns an empty string. The caller is
// responsible for creating a default file path if needed.
func (fp filePath) New() string {
	// temporarily storing values for the file path in an array allows for any
	// individual field to be used as the filename if no other fields are set.
	arr := []string{}

	// PrefixKey takes precedence over Prefix.
	switch {
	case fp.PrefixKey != "":
		arr = append(arr, "${PATH_PREFIX}")
	case fp.Prefix != "":
		arr = append(arr, fp.Prefix)
	}

	if fp.TimeFormat != "" {
		now := time.Now()

		// these options mirror process/time.go
		switch fp.TimeFormat {
		case "unix":
			arr = append(arr, fmt.Sprintf("%d", now.Unix()))
		case "unix_milli":
			arr = append(arr, fmt.Sprintf("%d", now.UnixMilli()))
		default:
			arr = append(arr, now.Format(fp.TimeFormat))
		}
	}

	if fp.UUID {
		arr = append(arr, uuid.NewString())
	}

	// SuffixKey takes precedence over Suffix.
	switch {
	case fp.SuffixKey != "":
		arr = append(arr, "${PATH_SUFFIX}")
	case fp.Suffix != "":
		arr = append(arr, fp.Suffix)
	}

	// if only one field is set, then this returns a filename, otherwise
	// it returns a file path.
	return path.Join(arr...)
}

// NewFileExtension returns a file extension based on file format and
// compression settings. The file extensions constructed by this function
// match this regular expression: `(\.json|\.txt)?(\.gz|\.zst)?`.
func NewFileExtension(format config.Config, compression config.Config) (ext string) {
	switch format.Type {
	case "data":
		break
	case "json":
		ext = ".json"
	case "text":
		ext = ".txt"
	}

	switch compression.Type {
	case "gzip":
		ext += ".gz"
	case "snappy":
		break
	case "zstd":
		ext += ".zst"
	}

	return ext
}
