module github.com/brexhq/substation/v2

go 1.23.0

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go-v2 v1.37.1
	github.com/aws/aws-sdk-go-v2/config v1.28.6
	github.com/aws/aws-sdk-go-v2/credentials v1.17.47
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.15.21
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression v1.7.56
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.43
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.42.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.38.0
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.36.0
	github.com/aws/aws-sdk-go-v2/service/firehose v1.35.2
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.32.7
	github.com/aws/aws-sdk-go-v2/service/lambda v1.69.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.85.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.34.7
	github.com/aws/aws-sdk-go-v2/service/sns v1.33.7
	github.com/aws/aws-sdk-go-v2/service/sqs v1.37.2
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.2
	github.com/aws/aws-xray-sdk-go/v2 v2.0.0
	github.com/aws/smithy-go v1.22.5
	github.com/awslabs/kinesis-aggregation/go/v2 v2.0.0-20241004223953-c2774b1ab29b
	github.com/golang/protobuf v1.5.4
	github.com/google/go-jsonnet v0.20.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/iancoleman/strcase v0.3.0
	github.com/itchyny/gojq v0.12.17
	github.com/klauspost/compress v1.17.11
	github.com/oschwald/maxminddb-golang v1.13.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.1
	github.com/tidwall/gjson v1.18.0 // Upgrades require SemVer bump.
	github.com/tidwall/sjson v1.2.5
	golang.org/x/exp v0.0.0-20241009180824-f66d83c29e7c
	golang.org/x/net v0.41.0
	golang.org/x/sync v0.16.0
)

require (
	cloud.google.com/go/storage v1.54.0
	github.com/GoogleCloudPlatform/functions-framework-go v1.9.2
	github.com/cloudevents/sdk-go/v2 v2.15.2
)

require (
	cel.dev/expr v0.20.0 // indirect
	cloud.google.com/go v0.121.0 // indirect
	cloud.google.com/go/auth v0.16.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/functions v1.19.3 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/monitoring v1.24.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.27.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.51.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.51.0 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.24.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.6 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250121191232-2f005788dc42 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.2 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.58.0 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.35.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	google.golang.org/api v0.232.0 // indirect
	google.golang.org/genproto v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250505200425-f936aa4a68b2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250505200425-f936aa4a68b2 // indirect
	google.golang.org/grpc v1.72.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
