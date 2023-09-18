# pubsub

This example deploys a data pipeline that uses a [publish/subscribe (pub/sub) pattern](https://aws.amazon.com/what-is/pub-sub-messaging/). The publisher receives change data capture (CDC) events from a DynamoDB table and publishes them to an SNS topic from which three subscribers consume them.

The data pipeline is visualized below:
```mermaid

flowchart TD
    %% core infrastructure
    ddb[DynamoDB Table]
    sns[SNS Topic]

    %% nodes
    publisher(Publisher)
    subscriber_x(Subscriber)
    subscriber_y(Subscriber)
    subscriber_z(Subscriber)

    %% connections
    ddb --> publisher
    publisher -->sns
    sns-->subscriber_x
    sns-->subscriber_y
    sns-->subscriber_z
```
