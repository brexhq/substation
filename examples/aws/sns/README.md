# sns

Contains example deployments that focus on AWS Simple Notification Service (SNS).

## pub_sub

Deploys a Substation data pipeline that implements a [publish/subscribe (pub/sub) pattern](https://aws.amazon.com/what-is/pub-sub-messaging/). DynamoDB Change Data Capture (CDC) events are published to and consumed from an SNS topic. 

The deployment is visualized below:
```mermaid

flowchart LR
    %% resources
    data[/Data/]
    ddb([DynamoDB Table])
    sns([SNS Topic])
    ingest[[Handler]]
    load[Transforms]
    subHandler1[[Handler]]
    subTransform1[Transforms]
    subHandler2[[Handler]]
    subTransform2[Transforms]
    subHandler3[[Handler]]
    subTransform3[Transforms]
    %% connections
    data --> ddb --> ingest
    subgraph Substation Publisher Node 
    ingest --> load
    end
    load --> sns 
    sns --> subHandler1
    sns --> subHandler2
    sns --> subHandler3
    subgraph Substation Subscriber Node 
    subHandler3 --> subTransform3
    end
    subgraph Substation Subscriber Node 
    subHandler2 --> subTransform2
    end
    subgraph Substation Subscriber Node 
    subHandler1 --> subTransform1
    end
```
