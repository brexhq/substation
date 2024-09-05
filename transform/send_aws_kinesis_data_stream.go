package transform

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	rec "github.com/awslabs/kinesis-aggregation/go/v2/records"
	"github.com/google/uuid"

	//nolint: staticcheck // not ready to switch package
	"github.com/golang/protobuf/proto"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// Records greater than 1 MB in size cannot be
// put into a Kinesis Data Stream.
const sendAWSKinesisDataStreamMessageSizeLimit = 1000 * 1000

// errSendAWSKinesisDataStreamMessageSizeLimit is returned when data
// exceeds the Kinesis record size limit. If this error occurs, then
// conditions or transforms should be applied to either drop or reduce
// the size of the data.
var errSendAWSKinesisDataStreamMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSKinesisDataStreamConfig struct {
	// UseBatchKeyAsPartitionKey determines if the batch key should be used as the partition key.
	UseBatchKeyAsPartitionKey bool `json:"use_batch_key_as_partition_key"`
	// EnableRecordAggregation determines if records should be aggregated.
	EnableRecordAggregation bool `json:"enable_record_aggregation"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
}

func (c *sendAWSKinesisDataStreamConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSKinesisDataStreamConfig) Validate() error {
	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSKinesisDataStream(ctx context.Context, cfg config.Config) (*sendAWSKinesisDataStream, error) {
	conf := sendAWSKinesisDataStreamConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_kinesis_data_stream: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_kinesis_data_stream"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendAWSKinesisDataStream{
		conf: conf,
	}

	// Kinesis Data Streams limits batch operations to 500 records.
	count := 500
	if conf.Batch.Count > 0 && conf.Batch.Count <= count {
		count = conf.Batch.Count
	}

	// Kinesis Data Streams limits batch operations to 5MiB.
	size := sendAWSKinesisDataStreamMessageSizeLimit * 5
	if conf.Batch.Size > 0 && conf.Batch.Size <= size {
		size = conf.Batch.Size
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    count,
		Size:     size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf.agg = agg

	if len(conf.AuxTransforms) > 0 {
		tf.tforms = make([]Transformer, len(conf.AuxTransforms))
		for i, c := range conf.AuxTransforms {
			t, err := New(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			tf.tforms[i] = t
		}
	}

	awsCfg, err := iconfig.NewAWS(ctx, conf.AWS)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.client = kinesis.NewFromConfig(awsCfg)

	return &tf, nil
}

type sendAWSKinesisDataStream struct {
	conf   sendAWSKinesisDataStreamConfig
	client *kinesis.Client

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSKinesisDataStream) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendAWSKinesisDataStreamMessageSizeLimit {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendAWSKinesisDataStreamMessageSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errBatchNoMoreData)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSKinesisDataStream) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSKinesisDataStream) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	var partitionKey string
	switch tf.conf.UseBatchKeyAsPartitionKey {
	case true:
		partitionKey = key
	case false:
		partitionKey = uuid.NewString()
	}

	if tf.conf.EnableRecordAggregation {
		data = tf.aggregateRecords(partitionKey, data)
	}

	if len(data) == 0 {
		return nil
	}

	ctx = context.WithoutCancel(ctx)
	if err := tf.putRecords(ctx, tf.conf.AWS.ARN, partitionKey, data); err != nil {
		return err
	}

	return nil
}

func (tf *sendAWSKinesisDataStream) aggregateRecords(partitionKey string, data [][]byte) [][]byte {
	var records [][]byte

	// Aggregation silently drops any data that is between ~0.9999 MB and 1 MB.
	agg := &sendAWSKinesisAggregate{}
	agg.New()

	for _, d := range data {
		if ok := agg.Add(d, partitionKey); ok {
			continue
		} else if agg.Count > 0 {
			records = append(records, agg.Get())
		}

		agg.New()
		_ = agg.Add(d, partitionKey)
	}

	if agg.Count > 0 {
		records = append(records, agg.Get())
	}

	return records
}

func (tf *sendAWSKinesisDataStream) putRecords(ctx context.Context, streamName, partitionKey string, data [][]byte) error {
	var entries []types.PutRecordsRequestEntry
	for _, d := range data {
		entries = append(entries, types.PutRecordsRequestEntry{
			Data:         d,
			PartitionKey: &partitionKey,
		})
	}

	resp, err := tf.client.PutRecords(ctx, &kinesis.PutRecordsInput{
		Records:   entries,
		StreamARN: &streamName,
	})
	if err != nil {
		return err
	}

	if resp.FailedRecordCount != nil && *resp.FailedRecordCount > 0 {
		var retry [][]byte

		for i, r := range resp.Records {
			if r.ErrorCode != nil {
				retry = append(retry, data[i])
			}
		}

		if len(retry) > 0 {
			return tf.putRecords(ctx, streamName, partitionKey, retry)
		}
	}

	return nil
}

// sendAWSKinesisAggregate produces a KPL-compliant Kinesis record
type sendAWSKinesisAggregate struct {
	Record       *rec.AggregatedRecord
	Count        int
	Size         int
	PartitionKey string
}

// New creates a new Kinesis record with default values
// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L167
func (a *sendAWSKinesisAggregate) New() {
	a.Record = &rec.AggregatedRecord{}
	a.Count = 0
	a.Size = 0

	a.PartitionKey = ""
	a.Record.PartitionKeyTable = make([]string, 0)
}

func varIntSize(i int) int {
	if i == 0 {
		return 1
	}

	var needed int
	for i > 0 {
		needed++
		i >>= 1
	}

	bytes := needed / 7
	if needed%7 > 0 {
		bytes++
	}

	return bytes
}

func (a *sendAWSKinesisAggregate) calculateRecordSize(data []byte, partitionKey string) int {
	var recordSize int
	// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L344-L349
	pkSize := 1 + varIntSize(len(partitionKey)) + len(partitionKey)
	recordSize += pkSize
	// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L362-L364
	pkiSize := 1 + varIntSize(a.Count)
	recordSize += pkiSize
	// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L371-L374
	dataSize := 1 + varIntSize(len(data)) + len(data)
	recordSize += dataSize
	// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L376-L378
	recordSize = recordSize + 1 + varIntSize(pkiSize+dataSize)

	// input record size + current aggregated record size + 4 byte magic header + 16 byte MD5 digest
	return recordSize + a.Record.XXX_Size() + 20
}

// Add inserts a Kinesis record into an aggregated Kinesis record
// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L382
func (a *sendAWSKinesisAggregate) Add(data []byte, partitionKey string) bool {
	// https://docs.aws.amazon.com/streams/latest/dev/key-concepts.html#partition-key
	if len(partitionKey) > 256 {
		partitionKey = partitionKey[0:256]
	}

	// grab the first parition key in the set of events
	if a.PartitionKey == "" {
		a.PartitionKey = partitionKey
	}

	// Verify the record size won't exceed the 1 MB limit of the Kinesis service.
	// https://docs.aws.amazon.com/streams/latest/dev/service-sizes-and-limits.html
	if a.calculateRecordSize(data, partitionKey) > 1024*1024 {
		return false
	}

	pki := uint64(a.Count)
	r := &rec.Record{
		PartitionKeyIndex: &pki,
		Data:              data,
	}

	// Append the data to the aggregated record.
	a.Record.Records = append(a.Record.Records, r)
	a.Record.PartitionKeyTable = append(a.Record.PartitionKeyTable, partitionKey)

	// Update the record count and size. This is not used in the aggregated record.
	a.Count++
	a.Size += a.calculateRecordSize(data, partitionKey)

	return true
}

// Get returns a KPL-compliant compressed Kinesis record
// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L293
func (a *sendAWSKinesisAggregate) Get() []byte {
	data, _ := proto.Marshal(a.Record)
	md5Hash := md5.Sum(data)

	record := []byte("\xf3\x89\x9a\xc2")
	record = append(record, data...)
	record = append(record, md5Hash[:]...)

	return record
}
