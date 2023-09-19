# sns

Contains example deployments that focus on AWS Simple Notification Service (SNS).

## pub_sub

Deploys a Substation data pipeline that implements a [publish/subscribe (pub/sub) pattern](https://aws.amazon.com/what-is/pub-sub-messaging/). DynamoDB Change Data Capture (CDC) events are published to and consumed from an SNS topic. 

The deployment is visualized below:
```mermaid

flowchart LR
    %% resources
    data[/Data/]
    kds1([Kinesis Data Stream])
    kds2([Kinesis Data Stream])

    pubHand[[Handler]]
    pubTform[Transforms]

    subHand[[Handler]]
    subTform[Transforms]

    %% connections
    data --> kds1 --> pubHand
    subgraph Substation Publisher Node 
    pubHand --> pubTform
    end

    pubTform --> kds2 --> subHand

    subgraph Substation Subscriber Node 
    subHand  --> subTform
    end
```
