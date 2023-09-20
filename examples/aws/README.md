# AWS

These example deployments demonstrate different use cases for Substation on AWS. 

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


# S3

## Data Lake

Deploys a data pipeline that implements a data lake by writing raw and processed data to an S3 bucket.

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

# SNS

## Pub/Sub

Deploys a data pipeline that implements a [publish/subscribe (pub/sub) pattern](https://aws.amazon.com/what-is/pub-sub-messaging/). The application in `cmd/development/substation` can act as the client by reading a local file and sending its content to the SNS topic.

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

Deploys an asynchronous microservice that performs DNS resolution. The service can be invoked [synchronously](https://docs.aws.amazon.com/lambda/latest/dg/invocation-sync.html) or using a [Lambda URL](https://docs.aws.amazon.com/lambda/latest/dg/lambda-urls.html) and results are stored in a DynamoDB table.

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
