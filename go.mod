module github.com/brexhq/substation/v2

go 1.22.0

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go v1.55.5
	github.com/aws/aws-sdk-go-v2 v1.32.2
	github.com/aws/aws-sdk-go-v2/config v1.28.0
	github.com/aws/aws-sdk-go-v2/credentials v1.17.41
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.15.12
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression v1.7.47
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.33
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.42.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.36.2
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.35.2
	github.com/aws/aws-sdk-go-v2/service/firehose v1.34.2
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.32.2
	github.com/aws/aws-sdk-go-v2/service/lambda v1.63.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.66.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.34.2
	github.com/aws/aws-sdk-go-v2/service/sns v1.33.2
	github.com/aws/aws-sdk-go-v2/service/sqs v1.36.2
	github.com/aws/aws-sdk-go-v2/service/sts v1.32.2
	github.com/aws/aws-xray-sdk-go v1.8.4
	github.com/awslabs/kinesis-aggregation/go/v2 v2.0.0-20241004223953-c2774b1ab29b
	github.com/golang/protobuf v1.5.4
	github.com/google/go-jsonnet v0.20.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/iancoleman/strcase v0.3.0
	github.com/itchyny/gojq v0.12.16
	github.com/klauspost/compress v1.17.11
	github.com/oschwald/maxminddb-golang v1.13.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.1
	github.com/tidwall/gjson v1.18.0 // Upgrades require SemVer bump.
	github.com/tidwall/sjson v1.2.5 // Upgrades require SemVer bump.
	golang.org/x/exp v0.0.0-20241009180824-f66d83c29e7c
	golang.org/x/net v0.30.0
	golang.org/x/sync v0.8.0
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.6 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.24.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.2 // indirect
	github.com/aws/smithy-go v1.22.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.56.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241021214115-324edc3d5d38 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
