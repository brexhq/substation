# dynamodb

Contains example deployments that focus on AWS DynamoDB.

## change data capture (cdc)

Deploys a Substation data pipeline that implements a [change data capture (CDC)](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html) pattern using DynamoDB Streams.

The deployment is visualized below:
```mermaid

flowchart LR
    %% resources
    data[/Data/]
    ddb([DynamoDB Table])

    cdcHandler[[Handler]]
    cdcTransforms[Transforms]

    %% connections
    data --> ddb --> cdcHandler
    subgraph Substation CDC Node 
    cdcHandler --> cdcTransforms
    end
```
