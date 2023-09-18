# s3

Contains example deployments that focus on AWS S3.

# microservice

Deploys a Substation data pipeline that implements a data lake by writing raw and processed data to an S3 bucket.

The service is visualized below:
```mermaid

flowchart LR
    data[/Data/]
    bucket([S3 Bucket])
    handler[[Handler]]
    gateway([API Gateway])

    sendS3x[Send to AWS S3]
    sendS3y[Send to AWS S3]
    mod[...]

    %% connections
    data --> gateway --> handler
    subgraph Substation Node
    handler --> sendS3x
    subgraph Transforms
    sendS3x --> mod --> sendS3y
    end
    end
    sendS3x --> bucket
    sendS3y --> bucket
```
