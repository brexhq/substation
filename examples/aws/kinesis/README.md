# kinesis

Contains example deployments that focus on AWS Kinesis.

## multistream

Deploys a Substation data pipeline that implements a multi-phase streaming data pattern using Kinesis Data Streams.

The deployment is visualized below:
```mermaid
flowchart LR
    %% resources
    data[/Data/]
    gateway([API Gateway])
    kds1([Kinesis Data Stream])
    kds2([Kinesis Data Stream])

    pubHand[[Handler]]
    pubTform[Transforms]

    subHand[[Handler]]
    subTform[Transforms]

    %% connections
    data --> gateway --> kds1 --> pubHand
    subgraph Substation Publisher Node 
    pubHand --> pubTform
    end

    pubTform --> kds2 --> subHand

    subgraph Substation Subscriber Node 
    subHand  --> subTform
    end
```
