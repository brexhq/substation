# sqs

Contains example deployments that focus on AWS SQS.

## microservice

Deploys Substation as an asynchronous microservice that performs DNS resolution. The service can be invoked [synchronously](https://docs.aws.amazon.com/lambda/latest/dg/invocation-sync.html) or using a [Lambda URL](https://docs.aws.amazon.com/lambda/latest/dg/lambda-urls.html) and results are stored in a DynamoDB table.

The deployment is visualized below:
```mermaid

flowchart LR
    %% resources
    gateway[HTTPS Endpoint]
    cli[AWS CLI]
    sqs([SQS Queue])
    ddb([DynamoDB Table])

    handler[[Handler]]
    tform[Transforms]

    serviceHand[[Handler]]
    serviceTform[Transforms]

    %% connections
    gateway <--> handler
    cli <--> handler

    subgraph Substation Frontend Node
    handler --> tform --> handler
    end

    tform --> sqs --> serviceHand

    subgraph Substation Microservice Node
    serviceHand --> serviceTform
    end

    serviceTform --> ddb
```
