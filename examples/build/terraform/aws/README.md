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

# Kinesis

## Firehose

Deploys a [Kinesis Data Firehose](https://aws.amazon.com/kinesis/data-firehose/) delivery stream with [data transformation](https://docs.aws.amazon.com/firehose/latest/dev/data-transformation.html) enabled.

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
