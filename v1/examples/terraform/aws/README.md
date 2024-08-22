# AWS

These example deployments demonstrate different use cases for Substation on AWS.

# CloudWatch Logs

## Cross-Account / Cross-Region

Deploys a data pipeline that collects data from CloudWatch log groups in any account or region into a Kinesis Data Stream.

```mermaid

flowchart LR
    %% resources
    cw1([CloudWatch Log Group])
    cw2([CloudWatch Log Group])
    cw3([CloudWatch Log Group])
    kds([Kinesis Data Stream])

    consumerHandler[[Handler]]
    consumerTransforms[Transforms]

    subgraph Account B / Region us-west-2
    cw2
    end

    subgraph Account A / Region us-west-2
    cw3
    end

    subgraph Account A / Region us-east-1
    cw1 --> kds
    cw3 --> kds
    cw2 --> kds
    kds --> consumerHandler

    subgraph Substation Consumer Node 
    consumerHandler  --> consumerTransforms
    end
    end
```

## To Lambda

Deploys a data pipeline that sends data from a CloudWatch log group to a Lambda function.

```mermaid

flowchart LR
    %% resources
    cw([CloudWatch Log Group])

    consumerHandler[[Handler]]
    consumerTransforms[Transforms]

    cw --> consumerHandler

    subgraph Substation Consumer Node 
    consumerHandler  --> consumerTransforms
    end
```

# DynamoDB

## Change Data Capture (CDC)

Deploys a data pipeline that implements a [change data capture (CDC) pattern using DynamoDB Streams](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html).

```mermaid

flowchart LR
    %% resources
    ddb([DynamoDB Table])

    cdcHandler[[Handler]]
    cdcTransforms[Transforms]

    %% connections
    ddb --> cdcHandler
    subgraph Substation CDC Node 
    cdcHandler --> cdcTransforms
    end
```

## Distributed Lock

Deploys a data pipeline that implements a distributed lock pattern using DynamoDB. This pattern can be used to add "exactly-once" semantics to services that otherwise do not support it. For similar examples, see the "exactly once" configurations [here](/examples/config/transform/meta/).

## Telephone

Deploys a data pipeline that implements a "telephone" pattern by sharing data as context between multiple Lambda functions using a DynamoDB table. This pattern can be used to enrich events across unique data sources.

```mermaid

flowchart LR
    %% resources
    md_kinesis([Device Management
    Kinesis Data Stream])
    edr_kinesis([EDR Kinesis Data Stream])
    idp_kinesis([IdP Kinesis Data Stream])
    dynamodb([DynamoDB Table])

    edrEnrichmentHandler[[Handler]]
    edrEnrichmentTransforms[Transforms]

    edrTransformHandler[[Handler]]
    edrTransformTransforms[Transforms]

    idpEnrichmentHandler[[Handler]]
    idpEnrichmentTransforms[Transforms]

    mdEnrichmentHandler[[Handler]]
    mdEnrichmentTransforms[Transforms]

    %% connections
    edr_kinesis --> edrEnrichmentHandler
    subgraph Substation EDR Enrichment Node 
    edrEnrichmentHandler --> edrEnrichmentTransforms
    end

    edr_kinesis --> edrTransformHandler
    subgraph Substation EDR Transform Node 
    edrTransformHandler --> edrTransformTransforms
    end

    idp_kinesis --> idpEnrichmentHandler
    subgraph Substation IdP Enrichment Node 
    idpEnrichmentHandler --> idpEnrichmentTransforms
    end

    md_kinesis --> mdEnrichmentHandler
    subgraph Substation Dvc Mgmt Enrichment Node 
    mdEnrichmentHandler --> mdEnrichmentTransforms
    end

    edrEnrichmentTransforms --- dynamodb
    edrTransformTransforms --- dynamodb
    idpEnrichmentTransforms --- dynamodb
    mdEnrichmentTransforms --- dynamodb
```

# EventBridge

## Lambda Bus

Deploys a data pipeline that sends data from an EventBridge event bus to a Lambda function.

```mermaid
flowchart LR
    %% resources
    ebb([EventBridge Bus])
    ebs([EventBridge Scheduler])

    producerHandler[[Handler]]
    producerTransforms[Transforms]

    consumerHandler[[Handler]]
    consumerTransforms[Transforms]

    %% connections
    ebs --> ebs
    ebs --> producerHandler
    subgraph Substation Producer Node 
    producerHandler --> producerTransforms
    end

    producerTransforms --> ebb --> consumerHandler

    subgraph Substation Consumer Node 
    consumerHandler  --> consumerTransforms
    end
```

# Firehose

## Data Transform

Deploys a [Firehose](https://aws.amazon.com/firehose/) delivery stream with [data transformation](https://docs.aws.amazon.com/firehose/latest/dev/data-transformation.html) enabled.

```mermaid

flowchart LR
    %% resources
    data[/Data/]
    firehose([Kinesis Data Firehose])
    s3([S3 Bucket])

    nodeHandler[[Handler]]
    nodeTransforms[Transforms]

    %% connections
    data --> firehose --> nodeHandler

    subgraph Substation Node
    nodeHandler --> nodeTransforms --> nodeHandler
    end

    nodeHandler --> firehose
    firehose --> s3
```

# Kinesis

## Autoscale

Deploys a Kinesis Data Stream with autoscaling enabled. This can also be used without Substation to manage Kinesis Data Streams.

```mermaid

flowchart LR
    kds[("Kinesis
    Data Stream")]
    sns("Autoscale SNS Topic")
    cw_upscale("CloudWatch Upscale Alarm")
    cw_downscale("CloudWatch Downscale Alarm")
    autoscale("Autoscale Lambda")

    autoscale -- UpdateShardCount API --> kds
    autoscale -- PutMetricAlarm API ---> cw_upscale
    autoscale -- PutMetricAlarm API ---> cw_downscale

    cw_downscale -. notifies .- sns
    cw_upscale -. notifies .- sns

    sns -- notifies ---> autoscale
    cw_downscale -. monitors .- kds
    cw_upscale -. monitors .- kds
```

## Multi-Stream

Deploys a data pipeline that implements a multi-phase streaming data pattern using Kinesis Data Streams.

```mermaid

flowchart LR
    %% resources
    gateway([API Gateway])
    kds1([Kinesis Data Stream])
    kds2([Kinesis Data Stream])

    publisherHandler[[Handler]]
    publisherTransforms[Transforms]

    subscriberHandler[[Handler]]
    subscriberTransforms[Transforms]

    %% connections
    gateway --> kds1 --> publisherHandler
    subgraph Substation Publisher Node 
    publisherHandler --> publisherTransforms
    end

    publisherTransforms --> kds2 --> subscriberHandler

    subgraph Substation Subscriber Node 
    subscriberHandler  --> subscriberTransforms
    end
```

## nXDR

Deploys a data pipeline that implements an nXDR pattern by applying threat / risk enrichment metadata to events and sending the enriched data to multiple destinations. This pattern is useful for:
- Generating risk-based detection rules
- Guiding analysts during incident investigations and incident response
- Aiding unstructured threat hunts
- Priorizing logs for retention and analysis

```mermaid

flowchart LR
    %% resources
    kinesis([Kinesis Data Stream])
    dynamodb([DynamoDB Table])
    ext([External System])

    enrichmentHandler[[Handler]]
    enrichmentTransforms[Transforms]

    transformHandler[[Handler]]
    transformTransforms[Transforms]

    %% connections
    kinesis --> enrichmentHandler
    subgraph Substation Enrichment Node 
    enrichmentHandler --> enrichmentTransforms
    end

    enrichmentTransforms --> dynamodb

    kinesis --> transformHandler
    subgraph Substation Transform Node 
    transformHandler --> transformTransforms
    end

    transformTransforms --> ext
```

## Time Travel

Deploys a data pipeline that implements a "time travel" pattern by having a subscriber node read data more slowly than an enrichment node. The nodes share data observed across different events using a DynamoDB table.

```mermaid

flowchart LR
    %% resources
    gateway([API Gateway])
    kinesis([Kinesis Data Stream])
    dynamodb([DynamoDB Table])

    enrichmentHandler[[Handler]]
    enrichmentTransforms[Transforms]

    subscriberHandler[[Handler]]
    subscriberTransforms[Transforms]

    gateway --> kinesis
    %% connections
    kinesis -- 5 seconds --> enrichmentHandler
    subgraph Substation Enrichment Node 
    enrichmentHandler --> enrichmentTransforms
    end

    enrichmentTransforms --> dynamodb

    kinesis -- 15 seconds --> subscriberHandler
    subgraph Substation Subscriber Node 
    subscriberHandler --> subscriberTransforms
    end

    dynamodb --- subscriberTransforms
```

# Lambda

## AppConfig

Deploys a data pipeline with an invalid config that triggers AppConfig's validator feature. When the AppConfig service receives the compiled Substation configuration and attempts to deploy, the deployment will fail and return an error.

## Microservice

Deploys a synchronous microservice that performs DNS resolution. The service can be invoked [synchronously](https://docs.aws.amazon.com/lambda/latest/dg/invocation-sync.html) or using a [Lambda URL](https://docs.aws.amazon.com/lambda/latest/dg/lambda-urls.html). 

```mermaid

flowchart LR
    %% resources
    gateway[HTTPS Endpoint]
    cli[AWS CLI]

    nodeHandler[[Handler]]
    nodeTransforms[Transforms]

    %% connections
    gateway <--> nodeHandler
    cli <--> nodeHandler

    subgraph Substation Node
    nodeHandler --> nodeTransforms --> nodeHandler
    end
```

## VPC

Deploys a synchronous microservice in a VPC that returns the public IP address of the Lambda function. The Lambda can be invoked [synchronously](https://docs.aws.amazon.com/lambda/latest/dg/invocation-sync.html) or using a [Lambda URL](https://docs.aws.amazon.com/lambda/latest/dg/lambda-urls.html). This example can be used to validate how Substation transforms behave inside a VPC.

```mermaid

flowchart LR
    %% resources
    gateway[HTTPS Endpoint]
    cli[AWS CLI]

    nodeHandler[[Handler]]
    nodeTransforms[Transforms]

    %% connections
    gateway <--> nodeHandler
    cli <--> nodeHandler

    subgraph VPC Network
    subgraph Substation Node
    nodeHandler --> nodeTransforms --> nodeHandler
    end
    end
```

# S3

## Data Lake

Deploys a data pipeline that implements a [data lake pattern using S3](https://docs.aws.amazon.com/whitepapers/latest/building-data-lakes/amazon-s3-data-lake-storage-platform.html). The S3 bucket contains two copies of the data (original and transformed).

```mermaid

flowchart LR
    bucket([S3 Bucket])
    handler[[Handler]]
    gateway([API Gateway])

    sendS3x[Send to AWS S3]
    sendS3y[Send to AWS S3]
    mod[...]

    %% connections
    gateway --> handler

    subgraph Substation Node
    handler --> sendS3x

    subgraph Transforms
    sendS3x --> mod --> sendS3y
    end

    end

    sendS3x --> bucket
    sendS3y --> bucket
```

## Retry on Failure

Deploys a data pipeline that reads data from an S3 bucket and automatically retries failed events using an SQS queue as a [failure destination](https://aws.amazon.com/blogs/compute/introducing-aws-lambda-destinations/). This example will retry forever until the error is resolved.

```mermaid

flowchart LR
    %% resources
    bucket([S3 Bucket])
    queue([SQS Queue])
    %% connections
    bucket --> handler
    N -.-> queue
    queue --> R
    rTransforms --> handler
    
    subgraph N["Substation Node"]
    handler[[Handler]] --> transforms[Transforms]
    end
    subgraph R["Substation Retrier"]
    rHandler[[Handler]] --> rTransforms[Transforms]
    end
```

## SNS

Deploys a data pipeline that reads data from an S3 bucket via an SNS topic.

```mermaid

flowchart LR
    %% resources
    bucket([S3 Bucket])
    sns([SNS Topic])

    handler[[Handler]]
    transforms[Transforms]

    %% connections
    bucket --> sns --> handler
    subgraph Substation Node 
    handler --> transforms
    end
```

## XDR

Deploys a data pipeline that implements an XDR (extended detection and response) pattern by reading files from an S3 bucket, conditionally filtering and applying threat / risk enrichment metadata to events, and then writing the enriched events to an S3 bucket. The S3 bucket contains two copies of the data (original and transformed).

```mermaid
flowchart LR
    bucket([S3 Bucket])
    handler[[Handler]]

    threat[Threat Enrichments]
    sendS3[Send to AWS S3]

    %% connections
    bucket --> handler

    subgraph Substation Node
    handler --> threat

    subgraph Transforms
    threat --> sendS3
    end

    end

    sendS3 --> bucket
```


# SNS

## Pub/Sub

Deploys a data pipeline that implements a [publish/subscribe (pub/sub) pattern](https://aws.amazon.com/what-is/pub-sub-messaging/). The `examples/cmd/client/file` application can be used to send data to the SNS topic.

```mermaid

flowchart LR
    %% resources
    file[(File)]
    sns([SNS Topic])

    cliHandler[[Handler]]
    cliTransforms[Transforms]
    sub1Handler[[Handler]]
    sub1Transforms[Transforms]
    sub2Handler[[Handler]]
    sub2Transforms[Transforms]
    sub3Handler[[Handler]]
    sub3Transforms[Transforms]

    %% connections
    cliHandler -.- file
    subgraph Substation Client 
    cliHandler --> cliTransforms
    end

    cliTransforms --> sns 
    sns --> sub1Handler
    sns --> sub2Handler
    sns --> sub3Handler

    subgraph Substation Subscriber Node 
    sub3Handler --> sub3Transforms
    end
    
    subgraph Substation Subscriber Node 
    sub2Handler --> sub2Transforms
    end
    
    subgraph Substation Subscriber Node 
    sub1Handler --> sub1Transforms
    end
```

# SQS

## Microservice

Deploys an asynchronous microservice that performs DNS resolution. The service can be invoked [synchronously](https://docs.aws.amazon.com/lambda/latest/dg/invocation-sync.html) or using a [Lambda URL](https://docs.aws.amazon.com/lambda/latest/dg/lambda-urls.html); requests to the service are assigned a UUID that can be used to retrieve the result from the DynamoDB table.

```mermaid

flowchart LR
    %% resources
    gateway[HTTPS Endpoint]
    cli[AWS CLI]
    sqs([SQS Queue])
    ddb([DynamoDB Table])

    gatewayHandler[[Handler]]
    gatewayTransforms[Transforms]

    microserviceHandler[[Handler]]
    microserviceTransforms[Transforms]

    %% connections
    gateway <--> gatewayHandler
    cli <--> gatewayHandler

    subgraph Substation Frontend Node
    gatewayHandler --> gatewayTransforms --> gatewayHandler
    end

    gatewayTransforms --> sqs --> microserviceHandler

    subgraph Substation Microservice Node
    microserviceHandler --> microserviceTransforms
    end

    microserviceTransforms --> ddb
```
